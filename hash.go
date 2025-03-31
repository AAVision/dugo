package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/zeebo/xxh3"
	"golang.org/x/crypto/blake2b"
)

const (
	hashThresholdMB = 20          // Files larger than this will use BLAKE2b
	bufferSize      = 32 * 1024
)

var bufPool = sync.Pool{
	New: func() interface{}{return make([]byte, bufferSize)},
}

func createFileHash(ctx context.Context, filePath string) (string, int64, error){
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file stats: %w", err)
	}

	fileSize := fileInfo.Size()

	sizeMB := float64(fileSize) / (1024 * 1024)
	var hash string

	if sizeMB > hashThresholdMB{
		hash, err = hashBlake2b(file)
	} else{
		hash, err = hashXXH3(file)
	}

	if err != nil {
		return "", fileSize, fmt.Errorf("hashing failed: %w", err)
	}

	return hash, fileSize, nil

}

func hashBlake2b(r io.Reader) (string, error){
	hasher, err := blake2b.New256(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create BLAKE2b hasher: %w", err)
	}

	buff := bufPool.Get().([]byte)
	defer bufPool.Put(buff)

	if _, err := io.CopyBuffer(hasher, r, buff); err != nil {
		return "", fmt.Errorf("BLAKE2b hashing failed: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashXXH3(r io.Reader) (string, error) {
	hasher := xxh3.New()
	buff := bufPool.Get().([]byte)
	defer bufPool.Put(buff)

	if _, err := io.CopyBuffer(hasher, r, buff); err != nil {
		return "", fmt.Errorf("XXH3 hashing failed: %w", err)
	}

	hash := hasher.Sum128()
	return fmt.Sprintf("%016x%016x", hash.Hi, hash.Lo), nil
}