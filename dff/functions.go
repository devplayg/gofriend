package dff

import (
	"encoding/hex"
	"fmt"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sort"
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

func isReadableDirs(dirs []string) ([]string, error) {
	absDirs := make([]string, 0)
	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}

		err = isValidDir(absDir)
		if err != nil {
			return nil, err
		}
		absDirs = append(absDirs, absDir)
	}

	return absDirs, nil
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
		for path, fileDetail := range filMapDetail.fileMap {
			fileMap[path] = fileDetail
		}
		log.Debugf("[%s] is merged into file map", filMapDetail.dir)
	}
	return fileMap, nil
}

func searchDir(dir string, minFileSize int64, ch chan *FileMapDetail) error {
	log.Infof("collecting files in [%s]", dir)
	fileMapDetail := NewFileMapDetail(dir)
	defer func() {
		//log.Debugf("finished searching files in [%s]", dir)
		ch <- fileMapDetail
	}()

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
			return err
		}
		if !f.IsDir() && f.Mode().IsRegular() && f.Size() >= minFileSize {
			fileMapDetail.fileMap[path] = &FileDetail{
				f:   f,
				dir: filepath.Dir(path),
			}
		}
		return nil
	})
	return err
}

func findDuplicateFiles(fileMap FileMap, minNumOfFilesInFileGroup int) (DuplicateFileMap, error) {
	fileMapBySize := classifyFilesBySize(fileMap)
	duplicateFileMap := make(DuplicateFileMap)
	for _, list := range fileMapBySize {
		if len(list) < minNumOfFilesInFileGroup {
			continue
		}

		updateDuplicateFileMap(duplicateFileMap, list)
	}
	return duplicateFileMap, nil
}

func classifyFilesBySize(fileMap FileMap) FileMapBySize {
	log.Debug("classifying files by size")
	fileMapBySize := make(FileMapBySize)
	for _, fileDetail := range fileMap {
		if _, ok := fileMapBySize[fileDetail.f.Size()]; !ok {
			fileMapBySize[fileDetail.f.Size()] = make([]*FileDetail, 0)
		}
		fileMapBySize[fileDetail.f.Size()] = append(fileMapBySize[fileDetail.f.Size()], fileDetail)
	}

	return fileMapBySize
}

func updateDuplicateFileMap(duplicateFileMap DuplicateFileMap, list []*FileDetail) {
	for _, fileDetail := range list {
		path := filepath.Join(fileDetail.dir, fileDetail.f.Name())
		key, err := generateFileKey(path)
		if err != nil {
			log.Error(err)
			continue
		}

		if _, ok := duplicateFileMap[key]; !ok {
			duplicateFileMap[key] = NewDuplicateFiles(fileDetail.f.Size())
		}
		duplicateFileMap[key].list = append(duplicateFileMap[key].list, path)
		duplicateFileMap[key].TotalSize += fileDetail.f.Size()
		duplicateFileMap[key].Count++
	}
}

func displayDuplicateFiles(duplicateFileMap DuplicateFileMap, totalFileCount int, minNumOfFilesInFileGroup int, sortBy int) {
	list := getSortedValues(duplicateFileMap, sortBy)
	no := 1
	for _, uniqFile := range list {
		if len(uniqFile.list) >= minNumOfFilesInFileGroup {
			fmt.Printf("file_no=#%d, size=%dB, count=%d, total_size=%s\n", no, uniqFile.Size, len(uniqFile.list), ByteCountDecimal(uniqFile.TotalSize))
			for _, path := range uniqFile.list {
				fmt.Printf("    - %s\n", path)
			}
			no++
		}
	}

	log.WithFields(log.Fields{
		"total_file_count":      totalFileCount,
		"duplicate_group_count": no - 1,
	}).Info("result")
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

// https://programming.guide/go/formatting-byte-size-to-human-readable-format.html
func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// https://programming.guide/go/formatting-byte-size-to-human-readable-format.html
func ByteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
