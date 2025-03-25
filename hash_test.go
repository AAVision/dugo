package main

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/blake2b"
)

func TestCreateFileHash(t *testing.T) {
	tempDir := t.TempDir()

	emptyFilePath := filepath.Join(tempDir, "empty.txt")
	emptyFile, err := os.Create(emptyFilePath)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	emptyFile.Close()

	contentFilePath := filepath.Join(tempDir, "content.txt")
	contentFile, err := os.Create(contentFilePath)
	if err != nil {
		t.Fatalf("Failed to create content test file: %v", err)
	}
	content := "Hello, world!"
	if _, err := contentFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
	contentFile.Close()

	h, err := blake2b.New256(nil)
	if err != nil {
		t.Fatalf("Failed to create BLAKE2b hash: %v", err)
	}
	n, err := io.WriteString(h, content)
	if err != nil {
		t.Error(err)
	}
	if n != len(content) {
		t.Errorf("not all bytes are written, expected to write %d bytes, written: %d", len(content), n)
	}
	expectedContentHash := hex.EncodeToString(h.Sum(nil))

	emptyHash, err := blake2b.New256(nil)
	if err != nil {
		t.Fatalf("Failed to create BLAKE2b hash for empty file: %v", err)
	}
	expectedEmptyHash := hex.EncodeToString(emptyHash.Sum(nil))

	tests := []struct {
		name           string
		filePath       string
		expectedHash   string
		expectedErrMsg string
	}{
		{
			name:           "Empty file",
			filePath:       emptyFilePath,
			expectedHash:   expectedEmptyHash,
			expectedErrMsg: "",
		},
		{
			name:           "File with content",
			filePath:       contentFilePath,
			expectedHash:   expectedContentHash,
			expectedErrMsg: "",
		},
		{
			name:           "Non-existent file",
			filePath:       filepath.Join(tempDir, "nonexistent.txt"),
			expectedHash:   "",
			expectedErrMsg: "failed to open file",
		},
		{
			name:           "Directory instead of file",
			filePath:       tempDir,
			expectedHash:   "",
			expectedErrMsg: "failed to compute hash",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := createFileHash(tc.filePath)

			if tc.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tc.expectedErrMsg)
				} else if msg := err.Error(); !contains(msg, tc.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedErrMsg, msg)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if hash != tc.expectedHash {
				t.Errorf("Expected hash %q, got %q", tc.expectedHash, hash)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
