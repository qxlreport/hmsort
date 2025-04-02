package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"

	"github.com/google/uuid"
)

func makeTestData(count int, dataFileName string) {

	fmt.Printf("make file %s with %d random 16 byte strings...\n", dataFileName, count)

	const PARTLEN = 1000000
	partCount := count / PARTLEN
	rest := count % PARTLEN

	data := make([]int32, PARTLEN)

	f, err := os.Create(dataFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 65*1024)

	num := 0

	write := func(datalen int) {
		for i := range datalen {
			data[i] = int32(num)
			num++
		}
		for i := range datalen {
			j := rand.Intn(i + 1)
			data[i], data[j] = data[j], data[i]
		}
		for i := range datalen {
			fmt.Fprintf(w, "%016d\n", data[i])
		}
	}

	for range partCount {
		write(PARTLEN)
	}
	if rest > 0 {
		write(rest)
	}
	w.Flush()
}

func makeTestDataUUID(count int, dataFileName string) {

	fmt.Printf("make file %s with %d 36 byte UUID strings...\n", dataFileName, count)

	f, err := os.Create(dataFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 65*1024)

	for range count {
		id := uuid.New()
		fmt.Fprintf(w, "%s\n", id.String())
	}

	w.Flush()
}
