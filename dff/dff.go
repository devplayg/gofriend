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
	sortBy            string
	accessDeniedCount int
}

func NewDuplicateFileFinder(dirs []string, minFileCount int, minFileSize int64) *DuplicateFileFinder {
	dff := DuplicateFileFinder{
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

	displayDuplicateFiles(duplicateFileMap, accessDeniedCount, d.minFileCount)

	return nil
}

//func (d *DuplicateFileFinder) findDuplicateFiles() error {
//
//	// Classify files by size
//	log.Debug("Classifying files by size")
//	fileMapBySize := make(FileMapBySize)
//	for _, fileDetail := range d.fileMap {
//		if _, ok := fileMapBySize[fileDetail.f.Size()]; !ok {
//			fileMapBySize[fileDetail.f.Size()] = make([]*FileDetail, 0)
//		}
//		fileMapBySize[fileDetail.f.Size()] = append(fileMapBySize[fileDetail.f.Size()], fileDetail)
//	}
//
//	duplicateFileMap := make(map[[32]byte]*DuplicateFiles)
//
//	for _, list := range fileMapBySize {
//
//		if len(list) < d.minFileCount {
//			continue
//		}
//
//		for _, fileDetail := range list {
//			path := filepath.Join(fileDetail.dir, fileDetail.f.Name())
//			key, err := generateFileKey(path)
//			if err != nil {
//				if strings.HasSuffix(err.Error(), "Access is denied.") {
//					d.accessDeniedCount++
//				} else {
//					log.Error(err)
//				}
//			}
//
//			if _, ok := duplicateFileMap[key]; !ok {
//				duplicateFileMap[key] = NewDuplicateFiles(fileDetail.f.Size())
//			}
//			duplicateFileMap[key].files = append(duplicateFileMap[key].files, path)
//
//		}
//	}
//
//	// Print
//	no := 1
//	for _, data := range duplicateFileMap {
//		totalSize := data.Size * int64(len(data.files))
//		//key.TotalSize = key.UnitSize * int64(len(list))
//		if len(data.files) > d.minFileCount {
//			fmt.Printf("no=#%d, unit_size=%d, count=%d, total_size=%d\n", no, data.Size, len(data.files), totalSize)
//			for _, path := range data.files {
//				fmt.Printf("    - %s\n", path)
//			}
//			no++
//		}
//	}
//	log.Infof("Access denied: %d", d.accessDeniedCount)
//
//	return nil
//}
