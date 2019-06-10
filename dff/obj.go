package dff

import (
	"errors"
	"hash"
	"os"
)

// Hash key
type Key struct {
	hash      [32]byte
	UnitSize  int64 // a file size
	TotalSize int64
}

type FileInfoDetail struct {
	dir string
	f   os.FileInfo
}

type FileMap map[string]*FileInfoDetail
type FileMapData struct {
	fileMap FileMap
	dir     string
}
type FileMapBySize map[int64][]*FileInfoDetail
type FileMapWhoseKeyIsHash map[Key][]string

// Errors
var (
	ErrAccessDenied = errors.New("access is denied")
)

var highwayHash hash.Hash
