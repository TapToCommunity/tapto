package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func GetMd5Hash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	_ = file.Close()
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GetFileSize(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return 0, err
	}

	size := stat.Size()
	_ = file.Close()

	return size, nil
}
