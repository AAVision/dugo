package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func main() {
	var ignoreNamesFlag, ignoreRegexFlag string
	var workers uint
	flag.StringVar(&ignoreNamesFlag, "ignore-names", "", "Comma-separated list of file/folder names to ignore (exact match)")
	flag.StringVar(&ignoreRegexFlag, "ignore-regex", "", "Regex pattern to ignore files by path")
	flag.UintVar(&workers, "workers", 4, "Number of concurrent workers")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("Usage: %s [options] <dir-path>", filepath.Base(os.Args[0]))
	}
	path := flag.Arg(0)

	ignoreNames := map[string]struct{}{}
	if ignoreNamesFlag != "" {
		for _, name := range strings.Split(ignoreNamesFlag, ",") {
			ignoreNames[name] = struct{}{}
		}
	}

	var ignoreRegex *regexp.Regexp
	if ignoreRegexFlag != "" {
		var err error
		ignoreRegex, err = regexp.Compile(ignoreRegexFlag)
		if err != nil {
			log.Fatalf("Invalid ignore regex: %v", err)
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	m, err := scanDir(abs, ignoreNames, ignoreRegex)
	if err != nil {
		log.Fatal(err)
	}

	sem := make(chan struct{}, workers)
	results := make(chan []string)

	var wg sync.WaitGroup
	for _, v := range m {
		if len(v) < 2 {
			continue
		}
		wg.Add(1)
		go func(files sameSizeFiles) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res, err := groupByHash(files, workers)
			if err != nil {
				log.Printf("Skipping due to error: %v", err)
				return
			}
			if len(res) < 2 {
				return
			}
			groups, err := partitionIntoEqualGroups(res)
			if err != nil {
				log.Printf("Error partitioning: %v", err)
				return
			}
			for _, group := range groups {
				if len(group) >= 2 {
					results <- group
				}
			}
		}(v)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for group := range results {
		fmt.Println("Equal files:", group)
	}

}
