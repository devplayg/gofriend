package dff

import (
	"encoding/hex"
	"fmt"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
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
	err := d.checkDirValidation()
	if err != nil {
		return err
	}

	// Collect all files
	ch := make(chan FileMapData, len(d.dirs))
	for _, dir := range d.dirs {
		go CollectFiles(dir, d.minFileSize, ch)
	}

	// Merge all map
	d.fileMap = make(FileMap)
	for i := 0; i < len(d.dirs); i++ {
		data := <-ch // Receive filemap from goroutine
		log.Debugf("got data from [%s]", data.dir)
		for path, fileInfoDetail := range data.fileMap {
			d.fileMap[path] = fileInfoDetail
		}
	}

	// Classify files by size
	log.Debug("Start comparing..")
	sizeMap := make(FileMapBySize)
	for _, fileDetail := range d.fileMap {
		if _, ok := sizeMap[fileDetail.f.Size()]; !ok {
			sizeMap[fileDetail.f.Size()] = make([]*FileInfoDetail, 0)
		}
		sizeMap[fileDetail.f.Size()] = append(sizeMap[fileDetail.f.Size()], fileDetail)
	}

	// Compare
	//mapWhoseKeyIsHash := map[string][]os.FileInfo
	//map whose key is
	//h := hash.Hash()

	//h.Sum()
	fileMapWhoseKeyIsHash := make(map[Key][]string)
	for _, list := range sizeMap {
		if len(list) > d.minFileCount {
			for _, fileDetail := range list {
				//f.Size()
				//spew.Dump(f)
				path := filepath.Join(fileDetail.dir, fileDetail.f.Name())
				_, err := ioutil.ReadFile(path)
				if err != nil {
					//
					switch err {
					case ErrAccessDenied:
						//log.Error("#err")
					default:
						if !strings.HasSuffix(err.Error(), "Access is denied.") {
							log.Error(err)
						}
					}
					continue
				}

				b, err := getHash(path)
				if err != nil {
					log.Error(err)
					continue
				}
				var hash [32]byte
				copy(hash[:], b)
				key := Key{
					hash:     hash,
					UnitSize: fileDetail.f.Size(),
				}
				if _, ok := fileMapWhoseKeyIsHash[key]; !ok {
					fileMapWhoseKeyIsHash[key] = make([]string, 0)
				}
				fileMapWhoseKeyIsHash[key] = append(fileMapWhoseKeyIsHash[key], path)
			}
		}
	}

	//spew.Dump(fileMapWhoseKeyIsHash)
	no := 1
	for key, list := range fileMapWhoseKeyIsHash {
		key.TotalSize = key.UnitSize * int64(len(list))
		if len(list) > d.minFileCount {
			fmt.Printf("no=#%d, unit_size=%d, count=%d, total_size=%d\n", no, key.UnitSize, len(list), key.TotalSize)
			for _, str := range list {
				fmt.Printf("    - %s\n", str)
			}
			no++
		}
	}

	return nil
}

func (d *DuplicateFileFinder) checkDirValidation() error {
	for _, dir := range d.dirs {
		err := IsValidDir(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func CollectFiles(dir string, minFileSize int64, ch chan FileMapData) error {
	data := FileMapData{
		fileMap: make(FileMap),
		dir:     dir,
	}
	log.Debugf("collecting files from directory [%s]", dir)
	defer func() {
		log.Debugf("done collecting files from [%s]", dir)
		ch <- data
	}()
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && f.Size() >= minFileSize {
			data.fileMap[path] = &FileInfoDetail{
				f:   f,
				dir: filepath.Dir(path),
			}
		}
		return nil
	})
	return err
}

func IsValidDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}
	return nil
}

func getHash(path string) ([]byte, error) {
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
