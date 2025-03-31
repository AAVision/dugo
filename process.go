package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
)

func groupByHash(ctx context.Context, files []string, workers uint) (map[string][]string, error) {
	type hashResult struct {
		hash string
		file string
		err  error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	hashChan := make(chan string, len(files))
	resultChan := make(chan hashResult, len(files))

	var wg sync.WaitGroup

	wg.Add(int(workers))

	for i := 0; i < int(workers); i++ {
		go func() {
			defer wg.Done()
			for file := range hashChan {
				select {
				case <-ctx.Done():
					return
				default:
					hash, _, err := createFileHash(ctx, file)
					resultChan <- hashResult{hash, file, err}
				}
			}
		}()
	}

	go func() {
		defer close(hashChan)

		for _, file := range files {
			select{
				case hashChan <- file:
				case <-ctx.Done():
					return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	m := make(map[string][]string)
	for res := range resultChan {
		if res.err != nil {
			cancel()
			return nil, res.err
		}
		m[res.hash] = append(m[res.hash], res.file)
	}

	mm := make(map[string][]string)
	for k, v := range m {
		if len(v) > 1 {
			mm[k] = v
		}
	}

	return mm, nil
}

func filesAreEqual(file1, file2 string) (bool, error) {
	f1, err := os.Open(file1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}
		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("read error: %v, %v", err1, err2)
		}
	}
}

func partitionIntoEqualGroups(files []string) ([][]string, error) {
	var groups [][]string
	for _, file := range files {
		matched := false
		for i, group := range groups {
			eq, err := filesAreEqual(group[0], file)
			if err != nil {
				return nil, err
			}
			if eq {
				groups[i] = append(groups[i], file)
				matched = true
				break
			}
		}
		if !matched {
			groups = append(groups, []string{file})
		}
	}
	return groups, nil
}
