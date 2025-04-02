package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qxlreport/hmsort"
)

func main() {

	dataFileName := "data.txt"
	outFileName := "sorted.txt"
	tempFileName := "temp.tmp"

	//if you need to see the content of a file, comment it defer os.Remove()
	defer os.Remove(dataFileName)
	defer os.Remove(tempFileName)
	defer os.Remove(outFileName)

	//make 20_000_000 data records
	if _, err := os.Stat(dataFileName); errors.Is(err, os.ErrNotExist) {
		makeTestData(1_000_000*20, dataFileName)
		//makeTestDataUUID(1_000_000*20, dataFileName)
	}

	var err error

	//open source file
	f, err := os.Open(dataFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	//buffered reader for fast read from dataFileName
	r := bufio.NewReaderSize(f, 65*1024)

	//open destination sorted file
	f2, err := os.Create(outFileName)
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	//buffered writer for fast save sorted data to outFileName
	writer := bufio.NewWriterSize(f2, 65*1024)

	//function for compare two records of data
	less := func(b1, b2 []byte) bool {
		return string(b1) < string(b2)
	}

	//function for get next record of data
	//should return []byte or nil, if no more data
	record := func() []byte {
		b, err := r.ReadBytes('\n')
		if err != nil {
			return nil
		}
		return b
	}

	fmt.Println("sort...")
	t := time.Now()

	//sort
	const sortBufferSize = 1024 * 1024 * 2
	err = hmsort.Sort(record, less, writer, tempFileName, sortBufferSize)
	if err != nil {
		panic(err)
	}

	//save sorted data to outFileName
	writer.Flush()
	t2 := time.Since(t)

	fmt.Println("ok")
	fmt.Printf("sort time: %v\n", t2)
}
