package dff

import (
	"encoding/hex"
	"github.com/minio/highwayhash"
)

func init() {
	key, err := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000")
	h, err := highwayhash.New(key)
	if err != nil {
		panic(err)
	}
	highwayHash = h
}

type DuplicateFileFinder struct {
	dirs              []string
	minFileCount      int
	minFileSize       int64
	sortBy            int
	accessDeniedCount int
}

func NewDuplicateFileFinder(dirs []string, minFileCount int, minFileSize int64, sortBy string) *DuplicateFileFinder {
	dff := DuplicateFileFinder{
		sortBy:       getSortValue(sortBy),
		dirs:         dirs,
		minFileCount: minFileCount,
		minFileSize:  minFileSize,
	}
	return &dff
}

func (d *DuplicateFileFinder) Start() error {
	err := isReadableDirs(d.dirs)
	if err != nil {
		return err
	}

	fileMap, err := collectFilesInDirs(d.dirs, d.minFileSize)
	if err != nil {
		return err
	}

	duplicateFileMap, accessDeniedCount, err := findDuplicateFiles(fileMap, d.minFileCount)
	if err != nil {
		return err
	}

	displayDuplicateFiles(duplicateFileMap, accessDeniedCount, d.minFileCount, d.sortBy)

	return nil
}
