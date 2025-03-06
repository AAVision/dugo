package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var ignoreNamesFlag, ignoreRegexFlag string
	flag.StringVar(&ignoreNamesFlag, "ignore-names", "", "Comma-separated list of file/folder names to ignore (exact match)")
	flag.StringVar(&ignoreRegexFlag, "ignore-regex", "", "Regex pattern to ignore files by path")
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
	for _, v := range m {
		if len(v) < 2 {
			continue
		}
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
				b, err := filesAreEqual(res[i], res[j])
				if err != nil {
					log.Fatal(err)
				}
				if b {
					eq[res[i]] = struct{}{}
					eq[res[j]] = struct{}{}
				}
			}
		}
		s := make([]string, 0, len(eq))
		for k := range eq {
			s = append(s, k)
		}
		fmt.Println("equal files:", s)
	}
}

type equalSet map[string]struct{}
