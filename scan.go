package main

import (
	"os"
	"path/filepath"
)

type sameSizeFiles []string

func scanDir(path string) (map[int64]sameSizeFiles, error) {
	m := make(map[int64]sameSizeFiles)
	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Type()&os.ModeSymlink != 0 { // if dir or symlink
			return nil
		}
		finfo, err := d.Info()
		if err != nil {
			return err
		}

		m[finfo.Size()] = append(m[finfo.Size()], path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}
