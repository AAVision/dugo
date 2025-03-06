package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func groupByHash(files []string) ([]string, error) {
	m := map[string][]string{}

	for _, v := range files {
		h, err := createFileHash(v)
		if err != nil {
			return nil, err
		}
		m[h] = append(m[h], v)
	}

	s := []string{}

	for _, v := range m {
		if len(v) > 1 {
			s = append(s, v...)
		}
	}

	return s, nil
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
