package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHandleDeletions(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tmpDir := t.TempDir()

	fileNames := []string{"file1.txt", "file2.txt", "file3.txt"}
	filePaths := make([]string, len(fileNames))
	for i, name := range fileNames {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		filePaths[i] = path
	}

	input := "1 3\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe creation failed: %v", err)
	}
	defer r.Close()

	_, err = w.WriteString(input)
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}
	w.Close()

	os.Stdin = r

	handleDeletions([][]string{filePaths})

	assertFileExists(t, filePaths[0], false)
	assertFileExists(t, filePaths[1], true)
	assertFileExists(t, filePaths[2], false)
}

func assertFileExists(t *testing.T, path string, shouldExist bool) {
	t.Helper()
	_, err := os.Stat(path)
	if shouldExist && os.IsNotExist(err) {
		t.Errorf("File %s should exist but was deleted", path)
	} else if !shouldExist && err == nil {
		t.Errorf("File %s should be deleted but exists", path)
	}
}

func TestParseDeleteInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		max         int
		expected    []int
		expectedErr bool
	}{
		{
			name:        "valid input with single number",
			input:       "1",
			max:         3,
			expected:    []int{0},
			expectedErr: false,
		},
		{
			name:        "valid input with multiple numbers",
			input:       "1 2 3",
			max:         3,
			expected:    []int{0, 1, 2},
			expectedErr: false,
		},
		{
			name:        "input with duplicates",
			input:       "1 2 2 3",
			max:         3,
			expected:    []int{0, 1, 2},
			expectedErr: false,
		},
		{
			name:        "input with invalid number",
			input:       "1 4",
			max:         3,
			expected:    nil,
			expectedErr: true,
		},
		{
			name:        "empty input",
			input:       "",
			max:         3,
			expected:    nil,
			expectedErr: false,
		},
		{
			name:        "input with out-of-range number",
			input:       "0 2",
			max:         3,
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDeleteInput(tt.input, tt.max)

			if tt.expectedErr && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !equal(result, tt.expected) {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDeleteFiles(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		files       []string
		indices     []int
		expectedErr bool
	}{
		{
			name:        "delete single file",
			files:       []string{"file1.txt"},
			indices:     []int{0},
			expectedErr: false,
		},
		{
			name:        "delete multiple files",
			files:       []string{"file1.txt", "file2.txt"},
			indices:     []int{0, 1},
			expectedErr: false,
		},
		{
			name:        "delete some files",
			files:       []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"},
			indices:     []int{1, 3},
			expectedErr: false,
		},
		{
			name:        "delete non-existent file",
			files:       []string{"file1.txt"},
			indices:     []int{0},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var group []string
			for _, file := range tt.files {
				filePath := filepath.Join(tempDir, file)
				group = append(group, filePath)
				if !tt.expectedErr {
					_, err := os.Create(filePath)
					if err != nil {
						t.Fatalf("Failed to create file: %v", err)
					}
				}
			}

			deleteFiles(group, tt.indices)

			for _, idx := range tt.indices {
				path := group[idx]
				_, err := os.Stat(path)
				if err == nil {
					t.Errorf("Failed to delete file %s: %v", path, err)
				}
			}
		})
	}
}
