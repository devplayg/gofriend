package dff

import (
	"encoding/hex"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
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
	dirs                     []string
	minNumOfFilesInFileGroup int
	minFileSize              int64
	sortBy                   int
	accessDeniedCount        int
}

func NewDuplicateFileFinder(dirs []string, minNumOfFilesInFileGroup int, minFileSize int64, sortBy string) *DuplicateFileFinder {
	dff := DuplicateFileFinder{
		sortBy:                   getSortValue(sortBy),
		dirs:                     dirs,
		minNumOfFilesInFileGroup: minNumOfFilesInFileGroup,
		minFileSize:              minFileSize,
	}
	log.WithFields(log.Fields{
		"min_file_size":                         minFileSize,
		"minimum_number_of_files_in_file_group": minNumOfFilesInFileGroup,
		"sort_by":                               sortBy,
	}).Info("settings")

	return &dff
}

func (d *DuplicateFileFinder) Init(verbose bool) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func (d *DuplicateFileFinder) Start() error {
	absDirs, err := isReadableDirs(d.dirs)
	if err != nil {
		return err
	}
	d.dirs = absDirs

	fileMap, err := collectFilesInDirs(d.dirs, d.minFileSize)
	if err != nil {
		return err
	}

	duplicateFileMap, err := findDuplicateFiles(fileMap, d.minNumOfFilesInFileGroup)
	if err != nil {
		return err
	}

	displayDuplicateFiles(duplicateFileMap, len(fileMap), d.minNumOfFilesInFileGroup, d.sortBy)
	return nil
}
