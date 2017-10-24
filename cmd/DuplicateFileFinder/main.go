package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	fs *flag.FlagSet
	fm *FileMap
	t1 time.Time
)

// File map
type FileMap struct {
	sync.RWMutex
	m map[string]*File
}

func NewFileMap() *FileMap {
	return &FileMap{
		m: make(map[string]*File),
	}
}
func (fm *FileMap) Load(key string) (value *File, ok bool) {
	fm.RLock()
	result, ok := fm.m[key]
	fm.RUnlock()
	return result, ok
}
func (fm *FileMap) Delete(key string) {
	fm.Lock()
	delete(fm.m, key)
	fm.Unlock()
}
func (fm *FileMap) Store(key string, path string, f os.FileInfo) {
	fm.Lock()
	result, ok := fm.m[key]
	if ok {
		result.count += 1
		result.totalSize += f.Size()
		result.list = append(result.list, path)
	} else {
		fm.m[key] = newFile(path, f.Size())
	}
	fm.Unlock()
}

// File
type File struct {
	size      int64
	totalSize int64
	count     int
	list      []string
}

func newFile(path string, size int64) *File {
	f := File{
		list:      make([]string, 0, 10),
		size:      size,
		totalSize: size,
		count:     1,
	}
	f.list = append(f.list, path)

	return &f
}

// Sort structure
type ValSorter struct {
	Keys []string
	Vals []*File
}

func NewValSorter(m map[string]*File) *ValSorter {
	vs := &ValSorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]*File, 0, len(m)),
	}
	for k, v := range m {
		vs.Keys = append(vs.Keys, k)
		vs.Vals = append(vs.Vals, v)
	}
	return vs
}
func (vs *ValSorter) Sort() {
	sort.Sort(sort.Reverse(vs))

}
func (vs *ValSorter) Len() int           { return len(vs.Vals) }
func (vs *ValSorter) Less(i, j int) bool { return vs.Vals[i].totalSize < vs.Vals[j].totalSize }
func (vs *ValSorter) Swap(i, j int) {
	vs.Vals[i], vs.Vals[j] = vs.Vals[j], vs.Vals[i]
	vs.Keys[i], vs.Keys[j] = vs.Keys[j], vs.Keys[i]
}

// Initialize
func init() {
	fm = NewFileMap()
	t1 = time.Now()
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// Main
func main() {
	const (
		Version = "1.0.1710.12401"
	)

	// Set and check flags
	fs = flag.NewFlagSet("", flag.ExitOnError)
	var (
		searchDir      = fs.String("d", "", "Source directory")
		countToDisplay = fs.Int("c", 3, "Minimum count")
		grCount        = fs.Int("gr", 300000, "Goroutine count")
		dispVer        = fs.Bool("v", false, "Print version")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])
	if *dispVer == true {
		fmt.Printf("dufind v%s\n", Version)
		return
	}
	if *searchDir == "" {
		printHelp()
		return
	}

	// Read and organize all files
	wg := new(sync.WaitGroup)
	c := make(chan bool, *grCount)
	var count int64
	var dispCount int64

	err := filepath.Walk(*searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			count += 1
			wg.Add(1)
			c <- true

			go func(name, path string, size int64) {
				s := strconv.FormatInt(size, 10)
				checksum := md5.Sum([]byte(name + s))
				key := hex.EncodeToString(checksum[:16])
				fm.Store(key, path, f)
				<-c
				wg.Done()
			}(f.Name(), path, f.Size())
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()

	// Sort
	vs := NewValSorter(fm.m)
	vs.Sort()

	// Print
	for idx, _ := range vs.Keys {
		if vs.Vals[idx].count >= *countToDisplay {
			fmt.Printf("# total %-15d bytes, each %-15d bytes, %d files\n", vs.Vals[idx].totalSize, vs.Vals[idx].size, vs.Vals[idx].count)
			for _, fn := range vs.Vals[idx].list {
				fmt.Printf("\t%s\n", fn)
			}
			dispCount += int64(vs.Vals[idx].count)
		}
	}
	fmt.Printf("\n# Time: %s, Count(displayed/total): %d/%d\n", time.Since(t1), dispCount, count)
}

func printHelp() {
	fmt.Println("dufinder [options]")
	fs.PrintDefaults()
}
