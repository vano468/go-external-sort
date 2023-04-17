package sorter

import (
	"bufio"
	"os"
)

type fileWrapper struct {
	path    string
	file    *os.File
	scanner *bufio.Scanner
}

func newFileWrapper(path string, file *os.File) *fileWrapper {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	return &fileWrapper{
		path:    path,
		file:    file,
		scanner: scanner,
	}
}

func (fw *fileWrapper) Scan() bool {
	return fw.scanner.Scan()
}

func (fw *fileWrapper) Text() string {
	return fw.scanner.Text()
}

func (fw *fileWrapper) File() *os.File {
	return fw.file
}

func (fw *fileWrapper) Clear() error {
	fw.file.Close()
	return os.Remove(fw.path)
}
