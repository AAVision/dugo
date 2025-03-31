package main

import (
	"context"
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
		n, err = io.WriteString(h, f.text)
		if err != nil {
			t.Error(err)
		}
		if n != len(f.text) {
			t.Errorf("not all bytes are written, expected to write %d bytes, written: %d", len(f.text), n)
		}
		f.hash = hex.EncodeToString(h.Sum(nil))
	}

	filesNames := make([]string, 0, len(files))
	for _, v := range files {
		filesNames = append(filesNames, filepath.Join(tempDir, v.file))
	}

	m, err := groupByHash(context.Background(), filesNames, workers)
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

func TestFilesAreEqual(t *testing.T) {
	tempDir := t.TempDir()

	files := [...]struct {
		name    string
		content string
	}{
		{"f1.txt", "hello file"},
		{"f2.txt", "hello file"},
		{"f3.txt", ""},
		{"f4.txt", ""},
		{"f5.txt", "HELLO FILE!"},
		{"f6.txt", "hello file!"},
		{"f7.txt", "hello world"},
		{"f8.txt", "hello world from Go"},
	}

	for _, f := range files {
		fullPath := filepath.Join(tempDir, f.name)
		file, err := os.Create(fullPath)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		n, err := io.WriteString(file, f.content)
		if err != nil {
			t.Error(err)
		}
		if n != len(f.content) {
			t.Errorf("not all bytes are written, expected to write: %d, written: %d", len(f.content), n)
		}
	}

	data := []struct {
		name      string
		file1     string
		file2     string
		equal     bool
		expectErr bool
	}{
		{"equal1", files[0].name, files[1].name, true, false},
		{"equal2", files[2].name, files[3].name, true, false},
		{"not-equal", files[3].name, files[1].name, false, false},
		{"not-equal-case-sensitive", files[4].name, files[5].name, false, false},
		{"not-equal-same-prefix", files[6].name, files[7].name, false, false},
		{"file doesn't exist", "Non-existent-file", files[0].name, false, true},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			f1 := filepath.Join(tempDir, d.file1)
			f2 := filepath.Join(tempDir, d.file2)
			b, err := filesAreEqual(f1, f2)
			if d.expectErr {
				if err == nil {
					t.Errorf("an error is expected, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Error(err)
			}
			if b != d.equal {
				t.Errorf("files are equal? expected: %v, got: %v", d.equal, b)
			}
		})
	}
}

func TestPartitionIntoEqualGroups(t *testing.T) {
	tempDir := t.TempDir()

	files := []struct {
		file string
		text string
	}{
		{"text1.txt", "hello text1"},
		{"text2.txt", "hello text1"},
		{"text3.txt", "hello text3"},
		{"text4.txt", "hello text3"},
		{"text5.txt", ""},
		{"text6.txt", ""},
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
	}

	filesNames := make([]string, 0, len(files))
	for _, v := range files {
		filesNames = append(filesNames, filepath.Join(tempDir, v.file))
	}

	groups, err := partitionIntoEqualGroups(filesNames)
	if err != nil {
		t.Error(err)
	}

	if len(groups) != 3 {
		t.Errorf("3 groups are expected, got: %d", len(groups))
	}

	for _, group := range groups {
		if len(group) != 2 {
			t.Errorf("group of length 2 is expected, got: %d", len(group))
		}
		for i := 0; i < len(files); i += 2 {
			f1 := filepath.Join(tempDir, files[i].file)
			f2 := filepath.Join(tempDir, files[i+1].file)

			if slices.Contains(group, f1) {
				if !slices.Contains(group, f2) {
					t.Errorf("files %q and %q are equal, but they are not in the same group", f1, f2)
				}
			}
		}
	}
}
