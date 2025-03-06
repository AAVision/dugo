package main

import (
	"os"
	"path/filepath"
	"regexp"
)

type sameSizeFiles []string

func scanDir(root string, ignoreNames map[string]struct{}, ignoreRegex *regexp.Regexp) (map[int64]sameSizeFiles, error) {
	filesBySize := make(map[int64]sameSizeFiles)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		base := filepath.Base(path)

		if _, ignored := ignoreNames[base]; ignored {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if ignoreRegex != nil && ignoreRegex.MatchString(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() || d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		finfo, err := d.Info()
		if err != nil {
			return err
		}

		filesBySize[finfo.Size()] = append(filesBySize[finfo.Size()], path)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return filesBySize, nil
}
