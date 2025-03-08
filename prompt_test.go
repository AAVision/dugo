package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
