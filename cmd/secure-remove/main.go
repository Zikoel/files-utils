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
	targetPath             string
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
	flag.StringVar(&c.targetPath, "target", "", "target folder")
	flag.StringVar(&c.targetPath, "t", "", "target folder (shorthand)")
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

	if c.targetPath == "" {
		flag.Usage()
		os.Exit(1)
	}
	absTargetPath, err := filepath.Abs(c.targetPath)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	c.targetPath = absTargetPath

	if c.bytesForHAsh < 0 {
		flag.Usage()
		os.Exit(1)
	}

	filesOnSource, err := listPathFiles(c.sourcePath, c.bytesForHAsh)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	filesOnTarget, err := listPathFiles(c.targetPath, c.bytesForHAsh)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	var possiblyFreeSpace int64 = 0
	var removableFilesCount int64 = 0
	var removableFiles = make([]firmedFile, 0)
	for _, file := range filesOnSource {
		fileDuplicate := fileExistInList(file, filesOnTarget)

		if fileDuplicate != nil {
			if c.verbosity == 0 {
				fmt.Printf("%s\n", file.path)
			} else if c.verbosity == 1 {
				fmt.Printf("You can delete file %s\n", file.path)
			} else if c.verbosity == 2 {
				fmt.Printf("You can delete file %s\n", file.path)
				fmt.Printf("\tduplicate on: \033[32m%s\033[0m\n", fileDuplicate.path)
			}
			possiblyFreeSpace += file.size
			removableFilesCount++
			removableFiles = append(removableFiles, file)
		}
	}

	if c.verbosity > 1 {
		fmt.Printf("%d\t\t files on source\n", len(filesOnSource))
		fmt.Printf("%d\t\t files on target\n", len(filesOnTarget))
		fmt.Printf("You can remove %d files\n", removableFilesCount)
		fmt.Printf("You can free %s\n", humanReadableSize(possiblyFreeSpace))
	}

	if c.generateDeleteCommands {
		fmt.Println("The comands to remove all files")
		for _, file := range removableFiles {
			fmt.Printf("rm '%s'\n", file.path)
		}
	}
}
