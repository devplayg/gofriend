package dff

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func getSortValue(sortBy string) int {
	var sortValue int
	sortBy = strings.TrimSpace(strings.ToLower(sortBy))
	switch sortBy {
	case "total":
		sortValue = SortByTotalSize
	case "size":
		sortValue = SortBySize
	case "count":
		sortValue = SortByCount
	default:
		sortValue = SortByTotalSize
	}

	return sortValue
}

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
		duplicateFileMap[key].list = append(duplicateFileMap[key].list, path)
		duplicateFileMap[key].TotalSize += fileDetail.f.Size()
		duplicateFileMap[key].Count++
	}
	return accessDeniedCount
}

func displayDuplicateFiles(duplicateFileMap DuplicateFileMap, accessDeniedCount, minFileCount int, sortBy int) {
	list := getSortedValues(duplicateFileMap, sortBy)
	no := 1
	for _, uniqFile := range list {
		if len(uniqFile.list) > minFileCount {
			fmt.Printf("no=#%d, unit_size=%d, count=%d, total_size=%d\n", no, uniqFile.Size, len(uniqFile.list), uniqFile.TotalSize)
			for _, path := range uniqFile.list {
				fmt.Printf("    - %s\n", path)
			}
			no++
		}
	}
}

func getSortedValues(duplicateFileMap DuplicateFileMap, sortBy int) []*UniqFile {
	list := make([]*UniqFile, 0, len(duplicateFileMap))
	for _, v := range duplicateFileMap {
		list = append(list, v)
	}

	// Sort by
	switch sortBy {
	case SortByCount:
		sort.Sort(ByCount{list})
	case SortBySize:
		sort.Sort(BySize{list})
	case SortByTotalSize:
		sort.Sort(ByTotalSize{list})
	default:
		sort.Sort(ByTotalSize{list})
	}

	return list
}
