# go-external-sort
This tool counts the occurrences of each row in a file. It's specifically designed for large files that cannot be loaded into memory all at once. You can set buffer size to ensure that no more than N elements are loaded into memory at any moment.

### Usage
```
go run main.go path/to/file <buffersize>
go run main.go example.txt 10
```

##### Input
```
test1
test2
test3
test2
test3
test2
test1
test3
test1
```

##### Output
```
test1	3
test2	3
test3	3
```

By default the tool outputs with [tsv](https://en.wikipedia.org/wiki/Tab-separated_values) file (output.tsv). You can change output format by providing different formatter func to the `WriteCounters`.

##### Limitations
At the current implementation, the tool keeps file descriptors opened for all temporary files per iteration. So if you have a too-large file and too-small buffer size, you may face with max open files limit per process or per login session (each OS has different limits). Feel free to provide a PR with changes to keep descriptors opened only for the current chunk 😊 (see [file_wrapper.go](https://github.com/vano468/go-external-sort/blob/main/sorter/file_wrapper.go)). Also as a workaround you may temporary [increase](https://www.tecmint.com/increase-set-open-file-limits-in-linux/) the limits.

### Testing
The repo contains two integration tests.
1. The integration test processing a small set of predefined data:
```
go test github.com/vano468/go-external-sort/sorter -v -run TestMergeSorterIntegration
```

2. The integration test generates and processes a large file containing 10M random numbers from 0 to 9. The test duration is about 60 seconds (m1 macbook) and it can be used to observe how the tool creates and manages chunks in the file system. Additionally, the test retains the output file, allowing for manual inspection of the final results, rather than relying on test checks:
```
go test github.com/vano468/go-external-sort/sorter -v -run TestMergeSorterLargeFile
```

### Algorithm
Underhood, the tool uses external merge sort algorithm:
1. Read the input file in chunks of size N, sort each chunk using a standard sorting algorithm and write to temporary file.
2. Take N sorted chunks from the temporary files and merge them using a min-heap of size N.
3. Continue merging the temporary files using the min-heap until they collapse into a single sorted file.

Once the input file is fully sorted, the algorithm performs a single scan to calculate the number of occurrences of each element and writes the results to the output file.
