package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Printf("usage: %s <dir-path>", args[0])
		os.Exit(1)
	}
	path := args[1]

	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	m, err := scanDir(abs)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range m {
		fmt.Printf("for size: %d, files: %v\n", k, v)
	}
}

func scanDir(path string) (map[int64][]string, error) {
	m := make(map[int64][]string)
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
