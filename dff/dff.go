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
	format                   int
}

func NewDuplicateFileFinder(dirs []string, minNumOfFilesInFileGroup int, minFileSize int64, sortBy string, format string) *DuplicateFileFinder {
	dff := DuplicateFileFinder{
		sortBy:                   getSortValue(sortBy),
		dirs:                     dirs,
		minNumOfFilesInFileGroup: minNumOfFilesInFileGroup,
		minFileSize:              minFileSize,
		format:                   getFormatValue(format),
	}

	return &dff
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

	duplicateFileGroupCount := d.displayDuplicateFileGroups(duplicateFileMap)

	log.WithFields(log.Fields{
		"number_of_files_scanned":               len(fileMap),
		"duplicate_group_count":                 duplicateFileGroupCount,
		"sort_by":                               d.sortBy,
		"minimum_number_of_files_in_file_group": d.minNumOfFilesInFileGroup,
		"min_file_size":                         d.minFileSize,
	}).Info("result")
	return nil
}

func (d *DuplicateFileFinder) displayDuplicateFileGroups(duplicateFileMap DuplicateFileMap) int {
	for key, uniqFile := range duplicateFileMap {
		if len(uniqFile.List) < d.minNumOfFilesInFileGroup {
			delete(duplicateFileMap, key)
		}
	}
	list := getSortedValues(duplicateFileMap, d.sortBy)
	duplicateFileGroupCount := len(list)

	if d.format == TextFormat {
		outputFilesInTextFormat(list)
		return duplicateFileGroupCount
	}
	outputFilesInJsonFormat(list)
	return duplicateFileGroupCount
}
