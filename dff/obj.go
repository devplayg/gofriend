package dff

import (
	"errors"
	"hash"
	"os"
)

// Errors
var (
	ErrAccessDenied = errors.New("access is denied")
)

var highwayHash hash.Hash

type FileDetail struct {
	dir string
	f   os.FileInfo
}

type FileMap map[string]*FileDetail

type FileMapDetail struct {
	fileMap FileMap
	dir     string
}

func NewFileMapDetail(dir string) *FileMapDetail {
	return &FileMapDetail{
		dir:     dir,
		fileMap: make(FileMap),
	}
}

type FileMapBySize map[int64][]*FileDetail

type DuplicateFiles struct {
	files []string
	Size  int64
}

func NewDuplicateFiles(size int64) *DuplicateFiles {
	return &DuplicateFiles{
		Size:  size,
		files: make([]string, 0),
	}
}

//type FileKey struct {
//	Hash  [32]byte
//	Size  int64
//	Count int
//}
//type ResultFileMap map[FileKey][]string
