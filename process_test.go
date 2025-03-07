package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGroupByHash(t *testing.T) {
	const workers = 4

	tempDir := t.TempDir()

	files := []struct {
		file string
		text string
		hash string
	}{
		{"text1.txt", "hello text1", ""},
		{"text2.txt", "hello text1", ""},
		{"text3.txt", "hello text3", ""},
		{"text4.txt", "hello text3", ""},
	}

	for _, f := range files {
		fullPath := filepath.Join(tempDir, f.file)
		file, err := os.Create(fullPath)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		n, err := io.WriteString(file, f.text)
		if err != nil {
			t.Error(err)
		}
		if n != len(f.text) {
			t.Errorf("not all bytes are written to %s, expected: %d, written: %d", f.file, len(f.text), n)
		}

		h := md5.New()
		io.WriteString(h, f.text)
		f.hash = hex.EncodeToString(h.Sum(nil))
	}

	filesNames := make([]string, 0, len(files))
	for _, v := range files {
		filesNames = append(filesNames, filepath.Join(tempDir, v.file))
	}

	m, err := groupByHash(filesNames, workers)
	if err != nil {
		t.Error(err)
	}

	if len(m) != 2 {
		t.Errorf("expected 2 hash groups, got: %d", len(m))
	}

	for k, v := range m {
		if len(v) != 2 {
			t.Errorf("expected a slice of length: %d, got: %d", 2, len(v))
		}
		if k == files[0].hash {
			if !slices.Contains(v, files[0].file) || !slices.Contains(v, files[1].file) {
				t.Errorf("files: %s and %s are expected to be in the %s hash group", files[0].file, files[1].file, k)
			}
		}
		if k == files[2].hash {
			if !slices.Contains(v, files[0].file) || !slices.Contains(v, files[1].file) {
				t.Errorf("files: %s and %s are expected to be in the %s hash group", files[0].file, files[1].file, k)
			}
		}
	}
}
