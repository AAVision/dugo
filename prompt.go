package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func handleDeletions(groups [][]string) {
	reader := bufio.NewReader(os.Stdin)
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		fmt.Printf("\nDuplicate group (%d files):\n", len(group))
		for i, path := range group {
			fmt.Printf("[%d] %s\n", i+1, path)
		}

		for {
			fmt.Print("Enter numbers to delete (space-separated, 'a' to abort): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if strings.ToLower(input) == "a" {
				break
			}

			toDelete, err := parseDeleteInput(input, len(group))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if len(toDelete) == len(group) {
				fmt.Println("Error: Cannot delete all files in group")
				continue
			}

			deleteFiles(group, toDelete)
			break
		}
	}
}

func parseDeleteInput(input string, max int) ([]int, error) {
	if input == "" {
		return nil, nil
	}

	seen := make(map[int]bool)
	var indices []int
	for _, s := range strings.Split(input, " ") {
		num, err := strconv.Atoi(s)
		if err != nil || num < 1 || num > max {
			return nil, fmt.Errorf("invalid number: %s", s)
		}
		if seen[num-1] {
			continue
		}
		seen[num-1] = true
		indices = append(indices, num-1)
	}
	return indices, nil
}

func deleteFiles(group []string, indices []int) {
	for _, idx := range indices {
		path := group[idx]
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("Failed to delete %s: %v\n", path, err)
		} else {
			fmt.Printf("Deleted: %s\n", path)
		}
	}
}
