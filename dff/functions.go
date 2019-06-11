package dff

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func isReadableDirs(dirs []string) error {
	for _, dir := range dirs {
		err := isValidDir(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func isValidDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}
	return nil
}

func generateFileKey(path string) ([32]byte, error) {
	hash, err := getHighwayHash(path)
	//logrus.Debugf("%s - %s", filepath.Base(path), hex.EncodeToString(hash))
	if err != nil {
		return [32]byte{}, err
	}

	var key [32]byte
	copy(key[:], hash)

	return key, nil
}

func getHighwayHash(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	highwayHash.Reset()
	if _, err = io.Copy(highwayHash, file); err != nil {
		return nil, err
	}

	checksum := highwayHash.Sum(nil)
	return checksum, nil
}

func collectFilesInDirs(dirs []string, minFileSize int64) (FileMap, error) {
	ch := make(chan *FileMapDetail, len(dirs))
	for _, dir := range dirs {
		go searchDir(dir, minFileSize, ch)
	}
	fileMap := make(FileMap)
	for i := 0; i < len(dirs); i++ {
		filMapDetail := <-ch // Receive filemap from goroutine
		log.Debugf("received files from [%s]", filMapDetail.dir)
		for path, fileDetail := range filMapDetail.fileMap {
			fileMap[path] = fileDetail
		}
	}
	return fileMap, nil
}

func searchDir(dir string, minFileSize int64, ch chan *FileMapDetail) error {
	log.Infof("searching files in [%s]", dir)
	fileMapDetail := NewFileMapDetail(dir)
	defer func() {
		log.Debugf("finished searching files in [%s]", dir)
		ch <- fileMapDetail
	}()
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

func findDuplicateFiles(fileMap FileMap, minFileCount int) (DuplicateFileMap, int, error) {

	var accessDeniedTotal int
	fileMapBySize := classifyFilesBySize(fileMap)
	duplicateFileMap := make(DuplicateFileMap)
	for _, list := range fileMapBySize {
		if len(list) < minFileCount {
			continue
		}

		accessDeniedCount := updateDuplicateFileMap(duplicateFileMap, list)
		accessDeniedTotal += accessDeniedCount
	}
	return duplicateFileMap, accessDeniedTotal, nil
}

func classifyFilesBySize(fileMap FileMap) FileMapBySize {
	log.Debug("Classifying files by size")
	fileMapBySize := make(FileMapBySize)
	for _, fileDetail := range fileMap {
		if _, ok := fileMapBySize[fileDetail.f.Size()]; !ok {
			fileMapBySize[fileDetail.f.Size()] = make([]*FileDetail, 0)
		}
		fileMapBySize[fileDetail.f.Size()] = append(fileMapBySize[fileDetail.f.Size()], fileDetail)
	}

	return fileMapBySize
}

func updateDuplicateFileMap(duplicateFileMap DuplicateFileMap, list []*FileDetail) int {
	var accessDeniedCount int
	for _, fileDetail := range list {
		path := filepath.Join(fileDetail.dir, fileDetail.f.Name())
		key, err := generateFileKey(path)
		if err != nil {
			if strings.HasSuffix(err.Error(), "Access is denied.") {
				accessDeniedCount++
			} else {
				log.Error(err)
			}
			continue
		}

		if _, ok := duplicateFileMap[key]; !ok {
			duplicateFileMap[key] = NewDuplicateFiles(fileDetail.f.Size())
		}
		duplicateFileMap[key].files = append(duplicateFileMap[key].files, path)
	}

	return accessDeniedCount
}

func displayDuplicateFiles(duplicateFileMap DuplicateFileMap, accessDeniedCount, minFileCount int) {
	no := 1
	for _, data := range duplicateFileMap {
		totalSize := data.Size * int64(len(data.files))
		//key.TotalSize = key.UnitSize * int64(len(list))
		if len(data.files) > minFileCount {
			fmt.Printf("no=#%d, unit_size=%d, count=%d, total_size=%d\n", no, data.Size, len(data.files), totalSize)
			for _, path := range data.files {
				fmt.Printf("    - %s\n", path)
			}
			no++
		}
	}
	log.Infof("Access denied files: %d", accessDeniedCount)
}
