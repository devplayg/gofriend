package dff

import (
	log "github.com/sirupsen/logrus"
)

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

//func (d *DuplicateFileFinder) Init(verbose bool) {
//	if verbose {
//		log.SetLevel(log.DebugLevel)
//	}
//}

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
