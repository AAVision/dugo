package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

func TestScanDir(t *testing.T) {
	tempDir := t.TempDir()

	testFiles := map[string]int{
		"file1.txt":               100,
		"file2.txt":               100,
		"file3.txt":               200,
		"ignored_file.txt":        300,
		"subdir/file4.txt":        400,
		"subdir/file5.txt":        400,
		"subdir/ignored_file.txt": 500,
		"regex_ignore/file6.txt":  600,
		"another_dir/file7.txt":   700,
		"symlink_target.txt":      800,
	}

	dirs := []string{
		filepath.Join(tempDir, "subdir"),
		filepath.Join(tempDir, "regex_ignore"),
		filepath.Join(tempDir, "another_dir"),
		filepath.Join(tempDir, "empty_dir"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	for filePath, size := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)

		file, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}

		if size > 0 {
			data := make([]byte, size)
			if _, err := file.Write(data); err != nil {
				file.Close()
				t.Fatalf("Failed to write to file %s: %v", fullPath, err)
			}
		}
		file.Close()
	}

	symlinkPath := filepath.Join(tempDir, "symlink.txt")
	symlinkTarget := filepath.Join(tempDir, "symlink_target.txt")
	if err := os.Symlink(symlinkTarget, symlinkPath); err != nil {
		// On Windows, creating symlinks may require admin privileges or developer mode
		// Skip this part if it fails
		t.Logf("Skipping symlink creation: %v", err)
	}

	tests := []struct {
		name          string
		root          string
		ignoreNames   map[string]struct{}
		ignoreRegex   *regexp.Regexp
		expectedSizes map[int64]int
		expectError   bool
	}{
		{
			name:        "Scan all files",
			root:        tempDir,
			ignoreNames: nil,
			ignoreRegex: nil,
			expectedSizes: map[int64]int{
				100: 2, // two files of size 100
				200: 1,
				300: 1,
				400: 2,
				500: 1,
				600: 1,
				700: 1,
				800: 1,
			},
			expectError: false,
		},
		{
			name:        "Ignore specific filename",
			root:        tempDir,
			ignoreNames: map[string]struct{}{"ignored_file.txt": {}},
			ignoreRegex: nil,
			expectedSizes: map[int64]int{
				100: 2,
				200: 1,
				400: 2,
				600: 1,
				700: 1,
				800: 1,
			},
			expectError: false,
		},
		{
			name:        "Ignore by regex",
			root:        tempDir,
			ignoreNames: nil,
			ignoreRegex: regexp.MustCompile(`regex_ignore`),
			expectedSizes: map[int64]int{
				100: 2,
				200: 1,
				300: 1,
				400: 2,
				500: 1,
				700: 1,
				800: 1,
			},
			expectError: false,
		},
		{
			name:        "Ignore both by name and regex",
			root:        tempDir,
			ignoreNames: map[string]struct{}{"ignored_file.txt": {}},
			ignoreRegex: regexp.MustCompile(`regex_ignore`),
			expectedSizes: map[int64]int{
				100: 2,
				200: 1,
				400: 2,
				700: 1,
				800: 1,
			},
			expectError: false,
		},
		{
			name:          "Non-existent directory",
			root:          filepath.Join(tempDir, "non_existent_dir"),
			ignoreNames:   nil,
			ignoreRegex:   nil,
			expectedSizes: nil,
			expectError:   true,
		},
		{
			name:        "Scan subdirectory only",
			root:        filepath.Join(tempDir, "subdir"),
			ignoreNames: nil,
			ignoreRegex: nil,
			expectedSizes: map[int64]int{
				400: 2,
				500: 1,
			},
			expectError: false,
		},
		{
			name:          "Empty directory",
			root:          filepath.Join(tempDir, "empty_dir"),
			ignoreNames:   nil,
			ignoreRegex:   nil,
			expectedSizes: map[int64]int{},
			expectError:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filesBySize, err := scanDir(tc.root, tc.ignoreNames, tc.ignoreRegex)

			if tc.expectError {
				if err == nil {
					t.Errorf("an error is expected, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Error(err)
			}

			sizeCounts := make(map[int64]int)
			for size, files := range filesBySize {
				sizeCounts[size] = len(files)
			}

			if len(tc.expectedSizes) != len(sizeCounts) {
				t.Error("File size counts don't match expected values")
			}
			for k, v := range tc.expectedSizes {
				if vv, ok := sizeCounts[k]; !ok || v != vv {
					t.Error("File size counts don't match expected values")
				}
			}

			for size, files := range filesBySize {
				if len(files) > 1 {
					for _, path := range files {
						stat, err := os.Stat(path)
						if err != nil {
							t.Errorf("error: %v", err)
						}
						if size != stat.Size() {
							t.Errorf("File %s has wrong size", path)
						}
					}

					for _, path := range files {
						_, err := os.Stat(path)
						if err != nil {
							t.Errorf("File path %s is not valid", path)
						}
					}
				}
			}

			for _, files := range filesBySize {
				for _, file := range files {
					fileInfo, err := os.Lstat(file)
					if err == nil && fileInfo.Mode()&os.ModeSymlink != 0 {
						t.Errorf("Symlink found in results: %s", file)
					}
				}
			}
		})
	}
}

// Utility function to normalize paths for consistent comparison across platforms
func normalizePaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for i, path := range paths {
		normalized[i] = filepath.ToSlash(path)
	}
	sort.Strings(normalized)
	return normalized
}
