package sorter

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"sort"
)

type MergeSorter struct {
	bufSize   int
	iteration int
	chunks    []*fileWrapper
	heap      *minHeap
}

func NewMergeSorter(bufSize int) *MergeSorter {
	hp := make(minHeap, 0, bufSize)
	return &MergeSorter{
		bufSize: bufSize,
		heap:    &hp,
	}
}

func (s *MergeSorter) Sort(input *os.File) (file *os.File, err error) {
	if s.chunks, err = s.createSortedChunks(input); err != nil {
		return
	}
	s.iteration++

	var nextChunks []*fileWrapper
	for len(s.chunks) > 1 {
		if nextChunks, err = s.mergeSortedChunks(s.chunks); err != nil {
			return
		}
		s.Clear()
		s.chunks = nextChunks
		s.iteration++
	}
	return s.chunks[0].file, nil
}

func (s *MergeSorter) WriteCounters(output *os.File, formatter func(str string, count int) string) error {
	var prev, cur string
	var occurences int

	var flush func() error
	flush = func() (err error) {
		if occurences > 0 {
			_, err = fmt.Fprintf(output, formatter(prev, occurences))
		}
		return
	}

	for s.chunks[0].Scan() {
		cur = s.chunks[0].Text()
		if cur == prev {
			occurences++
		} else {
			if err := flush(); err != nil {
				return err
			}
			prev = cur
			occurences = 1
		}
	}
	return flush()
}

func (s *MergeSorter) Clear() {
	for _, fw := range s.chunks {
		fw.Clear()
	}
}

func (s *MergeSorter) createSortedChunks(file *os.File) (chunks []*fileWrapper, err error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	buffer := make([]string, 0, s.bufSize)
	currentChunk := 1

	var flush func() error
	flush = func() error {
		sort.Strings(buffer)
		newChunk, err := s.writeToFile(currentChunk, func(file *os.File) error {
			for _, v := range buffer {
				if _, err := fmt.Fprintf(file, "%s\n", v); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		chunks = append(chunks, newChunk)
		currentChunk++
		buffer = buffer[:0]
		return nil
	}

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		if len(buffer) >= s.bufSize {
			if err = flush(); err != nil {
				return
			}
		}
		buffer = append(buffer, scanner.Text())
	}
	if len(buffer) > 0 {
		return chunks, flush()
	}
	return
}

func (s *MergeSorter) mergeSortedChunks(chunks []*fileWrapper) (mergedChunks []*fileWrapper, err error) {
	for from, currentChunk := 0, 1; from < len(chunks); from += s.bufSize {
		to := from + s.bufSize
		if to > len(chunks) {
			to = len(chunks)
		}

		for i := from; i < to; i++ {
			chunks[i].Scan()
			heap.Push(s.heap, heapEl{
				Str:     chunks[i].Text(),
				FileIdx: i,
			})
		}

		newChunk, err := s.writeToFile(currentChunk, func(file *os.File) error {
			for s.heap.Len() > 0 {
				el := heap.Pop(s.heap).(heapEl)
				if _, err := fmt.Fprintf(file, "%s\n", el.Str); err != nil {
					return err
				}
				if chunks[el.FileIdx].Scan() && chunks[el.FileIdx].Text() != "" {
					heap.Push(s.heap, heapEl{
						Str:     chunks[el.FileIdx].Text(),
						FileIdx: el.FileIdx,
					})
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		mergedChunks = append(mergedChunks, newChunk)
		currentChunk++
	}
	return
}

func (s *MergeSorter) writeToFile(chunk int, callback func(file *os.File) error) (*fileWrapper, error) {
	fileName := fmt.Sprintf(".tmp_%d_%d", s.iteration, chunk)
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	callback(file)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return newFileWrapper(fileName, file), nil
}
