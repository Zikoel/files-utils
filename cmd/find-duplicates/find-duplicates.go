package main

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type config struct {
	sourcePath             string
	bytesForHAsh           int64
	verbosity              int
	generateDeleteCommands bool
}

type firmedFile struct {
	path string
	size int64
	hash string
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func filesAreTheSame(file1, file2 firmedFile) bool {
	if file1.hash != file2.hash {
		return false
	}

	if file1.size != file2.size {
		return false
	}

	return true
}

func fileHash(folderPath string, info os.FileInfo, bytesForHash int64) (string, error) {

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

func listPathFiles(path string, bytesForHash int64) ([]firmedFile, error) {

	var result = make([]firmedFile, 0)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			innerFiles, err := listPathFiles(path+"/"+file.Name(), bytesForHash)
			if err != nil {
				return nil, err
			}
			result = append(result, innerFiles...)
		} else {
			hash, err := fileHash(path, file, bytesForHash)
			if err != nil {
				return nil, err
			}

			f := firmedFile{
				path: path + "/" + file.Name(),
				size: file.Size(),
				hash: hash,
			}
			result = append(result, f)
		}
	}

	return result, nil
}

func fileExistInList(file firmedFile, list []firmedFile) *firmedFile {
	for _, lf := range list {
		if filesAreTheSame(file, lf) {
			return &lf
		}
	}

	return nil
}

func humanReadableSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	}

	if size < 1024*1024 {
		var kb float64 = float64(size) / 1024
		return fmt.Sprintf("%f KB", kb)
	}

	if size < 1024*1024*1024 {
		var mb float64 = float64(size) / 1024 / 1024
		return fmt.Sprintf("%f MB", mb)
	}

	var gb float64 = float64(size) / 1024 / 1024 / 1024
	return fmt.Sprintf("%f GB", gb)
}

func main() {
	var c config

	flag.StringVar(&c.sourcePath, "source", "", "source folder")
	flag.StringVar(&c.sourcePath, "s", "", "source folder (shorthand)")
	flag.Int64Var(&c.bytesForHAsh, "b", 1024, "how many bytes of the file shoul be user for generate hash code 0 for all")
	flag.IntVar(&c.verbosity, "v", 0, "Verbosity level 0, 1, 2, 3")
	flag.BoolVar(&c.generateDeleteCommands, "g", false, "Generate deletion commands")

	flag.Parse()

	if c.sourcePath == "" {
		flag.Usage()
		os.Exit(1)
	}
	absSourcePath, err := filepath.Abs(c.sourcePath)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	c.sourcePath = absSourcePath

	if c.bytesForHAsh < 0 {
		flag.Usage()
		os.Exit(1)
	}

	filesOnSource, err := listPathFiles(c.sourcePath, c.bytesForHAsh)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	var duplicateMap = make(map[string][]firmedFile)

	for _, file := range filesOnSource {
		if duplicateMap[file.hash] == nil {
			duplicateMap[file.hash] = make([]firmedFile, 0)
		}
		duplicateMap[file.hash] = append(duplicateMap[file.hash], file)
	}

	if c.verbosity == 2 {
		fmt.Printf("Viewed files\t\t\t%d\n", len(filesOnSource))
		var duplicatedFiles int64 = 0
		var duplicatedSize int64 = 0
		for _, duplicates := range duplicateMap {
			if len(duplicates) > 1 {
				duplicatedFiles += int64(len(duplicates))
				duplicatedSize += duplicates[0].size * int64((len(duplicates) - 1))
			}

		}
		fmt.Printf("Duplicated files\t\t%d\n", duplicatedFiles)
		fmt.Printf("Wasted space files\t\t%s\n", humanReadableSize(duplicatedSize))
	}

	for hash, duplicates := range duplicateMap {
		if len(duplicates) > 1 {
			fmt.Printf("%s:\n", hash)
			for _, file := range duplicates {
				fmt.Printf("\t- %s\n", file.path)
			}
		}
	}
}
