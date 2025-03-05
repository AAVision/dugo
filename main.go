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
