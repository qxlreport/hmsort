package hmsort

import (
	"bufio"
	"container/heap"
	"encoding/binary"
	"io"
	"os"
	"sort"

	"codeberg.org/go-mmap/mmap"
)

const defaultPartBufferSize = 1024 * 1024 * 2
const writerBufferSize = 65 * 1024

// Function for compare two records of data.
// Should return true if b1 < b2
type Less func(b1, b2 []byte) bool

// Function for get next record of data.
// Should return []byte or nil if no more data
type Record func() []byte

// record - function for get next record of data.
// Should return []byte or nil if no more data
//
// less - function for compare two records of data.
// Should return true if b1 < b2
//
// outWriter - io.Writer for save sorted data
//
// tempFileName - name of temp file. File will be overwritten if it already exists
//
// partBufferSize - presort buffer size
func Sort(record Record, less Less, outWriter io.Writer, tempFileName string, partBufferSize int) (err error) {

	var temp *os.File
	temp, err = os.Create(tempFileName)
	if err != nil {
		return
	}
	temp.Close()

	temp, err = os.OpenFile(tempFileName, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer temp.Close()

	if partBufferSize == 0 {
		partBufferSize = defaultPartBufferSize
	}

	memBuf := make([]byte, partBufferSize)
	memPos := 0

	w := bufio.NewWriterSize(temp, writerBufferSize)

	maxRecSize := 0

	var ph []*partRec
	var sortList [][]byte
	var recBuf []byte
	var partOffset int64

	for {
		recBuf = record()
		L := len(recBuf)
		maxRecSize = max(maxRecSize, L)

		if recBuf == nil || memPos+L+4+4 > partBufferSize {
			if len(sortList) > 0 {
				sort.Slice(sortList, func(i, j int) bool { return less(sortList[i], sortList[j]) })
				for _, r := range sortList {
					_, err = w.Write(r)
					if err != nil {
						return
					}
				}
				_, err = w.Write([]byte{0, 0, 0, 0})
				if err != nil {
					return
				}
				sortList = sortList[:0]
				ph = append(ph, &partRec{offset: partOffset})
				partOffset += int64(memPos) + 4
				memPos = 0
			}
			if recBuf == nil {
				break
			}
		}
		r := memBuf[memPos : memPos+4+L]
		binary.LittleEndian.PutUint32(r[:4], uint32(L))
		copy(r[4:], recBuf)
		sortList = append(sortList, r)
		memPos += len(r)
	}
	err = w.Flush()
	if err != nil {
		return err
	}

	readRec := func(r io.Reader, buf *[]byte) error {
		*buf = (*buf)[:4]
		_, err = io.ReadFull(r, *buf)
		if err != nil {
			return err
		}
		recsize := binary.LittleEndian.Uint32((*buf)[:4])
		if recsize == 0 {
			return io.EOF
		}
		*buf = (*buf)[:recsize]
		_, err = io.ReadFull(r, *buf)
		if err != nil {
			return err
		}
		return nil
	}

	f, err := mmap.OpenFile(tempFileName, mmap.Read)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, part := range ph {
		_, err = f.Seek(part.offset, 0)
		if err != nil {
			return
		}

		part.buf = make([]byte, 4, maxRecSize)
		readRec(f, &part.buf)
		part.offset += int64(len(part.buf) + 4)
	}

	parts := partHeap{ph, less}
	heap.Init(&parts)

	for len(parts.ph) > 0 {
		p := heap.Pop(&parts)
		part := p.(*partRec)
		outWriter.Write(part.buf)

		f.Seek(part.offset, 0)
		if readRec(f, &part.buf) != io.EOF {
			heap.Push(&parts, part)
			part.offset += int64(len(part.buf) + 4)
		}
	}

	return nil
}
