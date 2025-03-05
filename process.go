package main

import (
	"bytes"
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

func compareByteByByte(file1, file2 string) (b bool, err error) {
	f1, err := os.Open(file1)
	if err != nil {
		return b, err
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return b, err
	}
	defer f2.Close()

	d1, err := io.ReadAll(f1)
	if err != nil {
		return b, err
	}
	d2, err := io.ReadAll(f2)
	if err != nil {
		return b, err
	}
	return bytes.Equal(d1, d2), nil
}
