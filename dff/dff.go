package dff

import (
	"encoding/hex"
	"fmt"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
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
	dirs         []string
	minFileCount int
	minFileSize  int64
	fileMap      FileMap
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

	// Check if directories are readable
	err := isReadableDirs(d.dirs)
	if err != nil {
		return err
	}

	// Collect all files
	ch := make(chan *FileMapDetail, len(d.dirs))
	for _, dir := range d.dirs {
		go CollectFiles(dir, d.minFileSize, ch)
	}
	d.fileMap = make(FileMap)
	for i := 0; i < len(d.dirs); i++ {
		filMapDetail := <-ch // Receive filemap from goroutine
		log.Debugf("received result from [%s]", filMapDetail.dir)
		for path, fileDetail := range filMapDetail.fileMap {
			d.fileMap[path] = fileDetail
		}
	}

	err = d.findDuplicateFiles()
	if err != nil {
		return err
	}

	return nil
}

func (d *DuplicateFileFinder) findDuplicateFiles() error {

	// Classify files by size
	log.Debug("Classifying files by size")
	fileMapBySize := make(FileMapBySize)
	for _, fileDetail := range d.fileMap {
		if _, ok := fileMapBySize[fileDetail.f.Size()]; !ok {
			fileMapBySize[fileDetail.f.Size()] = make([]*FileDetail, 0)
		}
		fileMapBySize[fileDetail.f.Size()] = append(fileMapBySize[fileDetail.f.Size()], fileDetail)
	}

	duplicateFileMap := make(map[[32]byte]*DuplicateFiles)

	for _, list := range fileMapBySize {

		if len(list) < d.minFileCount {
			continue
		}

		for _, fileDetail := range list {
			path := filepath.Join(fileDetail.dir, fileDetail.f.Name())
			key, err := generateFileKey(path)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "Access is denied.") {
					log.Warn(err)
				} else {
					log.Error(err)
				}
			}

			//if err != nil {
			//    log.Error(err)
			//    continue
			//}

			if _, ok := duplicateFileMap[key]; !ok {
				duplicateFileMap[key] = NewDuplicateFiles(fileDetail.f.Size())
			}
			duplicateFileMap[key].files = append(duplicateFileMap[key].files, path)

		}
	}

	no := 1
	for _, data := range duplicateFileMap {
		totalSize := data.Size * int64(len(data.files))
		//key.TotalSize = key.UnitSize * int64(len(list))
		if len(data.files) > d.minFileCount {
			fmt.Printf("no=#%d, unit_size=%d, count=%d, total_size=%d\n", no, data.Size, len(data.files), totalSize)
			for _, path := range data.files {
				fmt.Printf("    - %s\n", path)
			}
			no++
		}
	}
	return nil

}

func CollectFiles(dir string, minFileSize int64, ch chan *FileMapDetail) error {
	log.Debugf("collecting files from [%s]", dir)
	fileMapDetail := NewFileMapDetail(dir)
	defer func() {
		log.Debugf("finished collecting files from [%s]", dir)
		ch <- fileMapDetail
	}()

	// Collecting files
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && f.Size() >= minFileSize {
			fileMapDetail.fileMap[path] = &FileDetail{
				f:   f,
				dir: filepath.Dir(path),
			}
		}
		return nil
	})
	return err
}
