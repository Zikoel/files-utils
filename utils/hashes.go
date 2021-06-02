package utils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
)

// FileHash get the hash of a file
func FileHash(folderPath string, info os.FileInfo, bytesForHash int64) (string, error) {

	headerSize := min(bytesForHash, info.Size())

	r, err := os.Open(folderPath + "/" + info.Name())
	if err != nil {
		return "", err
	}
	defer r.Close()

	header := make([]byte, headerSize)
	n, err := io.ReadFull(r, header[:])
	if err != nil {
		return "", err
	}

	if int64(n) < headerSize {
		return "", errors.New("Not all the specified bytes can be readed")
	}

	md5 := fmt.Sprintf("%x", md5.Sum(header))

	return md5, nil
}
