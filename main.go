package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/vano468/go-external-sort/sorter"
)

const (
	outputPath = "output.tsv"
	minBufSize = 2
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Provide two args: input file and buffer size")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	bufSize, err := strconv.ParseInt(os.Args[2], 10, 0)
	if err != nil || bufSize < minBufSize {
		fmt.Println("Buffer size must be integer and >= 2")
		os.Exit(1)
	}

	mSorter := sorter.NewMergeSorter(int(bufSize))
	defer mSorter.Clear()

	inputFile, err := os.Open(inputPath)
	handleError(err)
	defer inputFile.Close()
	_, err = mSorter.Sort(inputFile)
	handleError(err)

	outputFile, err := os.Create(outputPath)
	handleError(err)
	defer outputFile.Close()
	handleError(mSorter.WriteCounters(outputFile, func(str string, count int) string {
		return fmt.Sprintf("%s\t%d\n", str, count)
	}))
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
