package sorter_test

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/vano468/go-external-sort/sorter"
)

func TestMergeSorterIntegration(t *testing.T) {
	bufSize := 2

	inputFile, cleanInputFile := createTestFileWithContent([]string{
		"this", "test", "asd", "the", "end", "sad", "this", "is", "asd", "end", "my", "only", "test", "the",
	})
	defer cleanInputFile()
	mSorter := sorter.NewMergeSorter(bufSize)
	defer mSorter.Clear()

	sortedFile, err := mSorter.Sort(inputFile)
	if err != nil {
		t.Fatalf("Error occurred while sorting: %v", err)
	}
	if !isFileSorted(sortedFile) {
		t.Fatal("Generated file is not sorted properly")
	}

	expectedCounters := map[string]int{
		"this": 2,
		"test": 2,
		"asd":  2,
		"the":  2,
		"end":  2,
		"sad":  1,
		"is":   1,
		"my":   1,
		"only": 1,
	}
	sortedFile.Seek(0, io.SeekStart)
	outputFile, cleanOutputFile := createTestFile(nil)
	defer cleanOutputFile()

	totalCounts := 0
	err = mSorter.WriteCounters(outputFile, func(str string, count int) string {
		totalCounts++
		if expectedCounters[str] != count {
			t.Errorf("Expected %d occurences for '%s', but got %d", expectedCounters[str], str, count)
		}
		return str
	})
	if err != nil {
		t.Fatalf("Error occurred while writing counters: %v", err)
	}
	if totalCounts != len(expectedCounters) {
		t.Errorf("Some rows are missed in resulted file, expected %d, got %d total count", len(expectedCounters), totalCounts)
	}
}

func TestMergeSorterLargeFile(t *testing.T) {
	bufSize := 50000
	fileRows := 10000000
	randLimit := 10

	inputFile, cleanInputFile := createTestFile(func(f *os.File) {
		for i := 0; i < fileRows; i++ {
			fmt.Fprintf(f, "%d\n", rand.Intn(randLimit))
		}
	})
	defer cleanInputFile()
	mSorter := sorter.NewMergeSorter(bufSize)
	defer mSorter.Clear()

	sortedFile, err := mSorter.Sort(inputFile)
	if err != nil {
		t.Fatalf("Error occurred while sorting: %v", err)
	}
	if !isFileSorted(sortedFile) {
		t.Fatal("Generated file is not sorted properly")
	}

	outputFile, _ := createTestFile(nil)
	sortedFile.Seek(0, io.SeekStart)
	totalCounts := 0
	err = mSorter.WriteCounters(outputFile, func(str string, count int) string {
		totalCounts++
		return fmt.Sprintf("%s\t%d\n", str, count)
	})
	if err != nil {
		t.Fatalf("Error occurred while writing counters: %v", err)
	}
	if totalCounts != randLimit {
		t.Errorf("Some rows are missed in resulted file, expected %d, got %d total count", randLimit, totalCounts)
	}
}

func createTestFileWithContent(content []string) (file *os.File, clean func()) {
	return createTestFile(func(f *os.File) {
		for _, str := range content {
			fmt.Fprintf(f, "%s\n", str)
		}
	})
}

func createTestFile(handler func(*os.File)) (file *os.File, clean func()) {
	fileName := fmt.Sprintf(".test_%d", time.Now().Unix())
	file, _ = os.Create(fileName)
	if handler != nil {
		handler(file)
	}
	file.Seek(0, io.SeekStart)
	return file, func() {
		file.Close()
		os.Remove(fileName)
	}
}

func isFileSorted(f *os.File) bool {
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var prev, cur string
	sorted := true
	for scanner.Scan() {
		cur = scanner.Text()
		if prev > cur {
			sorted = false
		}
	}
	return sorted
}
