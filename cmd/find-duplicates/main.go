package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/zikoel/files_utils"
	"github.com/zikoel/files_utils/utils"
)

type config struct {
	sourcePath             string
	bytesForHAsh           int64
	verbosity              int
	generateDeleteCommands bool
}

type FileFingerprint = files_utils.FileFingerprint

func listPathFiles(path string, bytesForHash int64) ([]FileFingerprint, error) {

	var result = make([]FileFingerprint, 0)

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
			hash, err := utils.FileHash(path, file, bytesForHash)
			if err != nil {
				return nil, err
			}

			f := FileFingerprint{
				Path: path + "/" + file.Name(),
				Size: file.Size(),
				Hash: hash,
			}
			result = append(result, f)
		}
	}

	return result, nil
}

func fileExistInList(file FileFingerprint, list []FileFingerprint) *FileFingerprint {
	for _, lf := range list {
		if utils.FilesAreTheSame(file, lf) {
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

	var duplicateMap = make(map[string][]FileFingerprint)

	for _, file := range filesOnSource {
		if duplicateMap[file.Hash] == nil {
			duplicateMap[file.Hash] = make([]FileFingerprint, 0)
		}
		duplicateMap[file.Hash] = append(duplicateMap[file.Hash], file)
	}

	if c.verbosity == 2 {
		fmt.Printf("Viewed files\t\t\t%d\n", len(filesOnSource))
		var duplicatedFiles int64 = 0
		var duplicatedSize int64 = 0
		for _, duplicates := range duplicateMap {
			if len(duplicates) > 1 {
				duplicatedFiles += int64(len(duplicates))
				duplicatedSize += duplicates[0].Size * int64((len(duplicates) - 1))
			}

		}
		fmt.Printf("Duplicated files\t\t%d\n", duplicatedFiles)
		fmt.Printf("Wasted space files\t\t%s\n", humanReadableSize(duplicatedSize))
	}

	for hash, duplicates := range duplicateMap {
		if len(duplicates) > 1 {
			fmt.Printf("%s:\n", hash)
			for _, file := range duplicates {
				fmt.Printf("\t- %s\n", file.Path)
			}
		}
	}
}
