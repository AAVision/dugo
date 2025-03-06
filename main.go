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
		eq := make(equalSet)
		for i := range res {
			for j := i + 1; j < len(res); j++ {
				b, err := compareByteByByte(res[i], res[j])
				if err != nil {
					log.Fatal(err)
				}
				if b {
					eq[res[i]] = struct{}{}
					eq[res[j]] = struct{}{}
				}
			}
		}
		s := make([]string, len(eq))
		for k := range eq {
			s = append(s, k)
		}
		fmt.Println("equal files:", s)
	}
}

type equalSet map[string]struct{}
