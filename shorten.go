package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func shorten(url string) (string, error) {
	hasher := md5.New()
	_, err := io.WriteString(hasher, url)
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
