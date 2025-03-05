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
	for _, v := range m {
		res, err := groupByHash(v)
		if err != nil {
			log.Fatal(err)
		}
		if len(res) < 2 {
			continue
		}
		fmt.Printf("res: %v\n", res)
		b, err := compareByteByByte(res[0], res[1])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("bytes are equal:", b, "res:", res)
	}
}
