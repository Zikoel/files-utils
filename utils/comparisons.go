package utils

import "github.com/zikoel/files_utils"

// FilesAreTheSame compare two firmed files
func FilesAreTheSame(file1, file2 files_utils.FileFingerprint) bool {
	if file1.Hash != file2.Hash {
		return false
	}

	if file1.Size != file2.Size {
		return false
	}

	return true
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
