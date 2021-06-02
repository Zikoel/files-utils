package files_utils

// FirmedFile a descriptor of a file. We call it firmed file
type FileFingerprint struct {
	Path string
	Size int64
	Hash string
}
