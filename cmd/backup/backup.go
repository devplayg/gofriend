package main

import (
	"fmt"
	"flag"
	"os"
)

var	fs *flag.FlagSet
const (
	DefaultHistoryFile = "/tmp/backup.dbb"
	Default = "/tmp/backup.dbb"
)


func main(){
	var (
		historyFile         = fs.String("repo", DefaultHistoryFile, "History file")
		srcDir = fs.String("s", "", "Source directory")
		dstDir = fs.String("d", "", "Destination directory")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])
}
//
//// File map
//type FileMap struct {
//	sync.RWMutex
//	m map[string]*File
//}
//
//func NewFileMap() *FileMap {
//	return &FileMap{
//		m: make(map[string]*File),
//	}
//}
//func (fm *FileMap) Load(key string) (value *File, ok bool) {
//	fm.RLock()
//	result, ok := fm.m[key]
//	fm.RUnlock()
//	return result, ok
//}
//func (fm *FileMap) Delete(key string) {
//	fm.Lock()
//	delete(fm.m, key)
//	fm.Unlock()
//}
//func (fm *FileMap) Store(key string, path string, f os.FileInfo) {
//	fm.Lock()
//	result, ok := fm.m[key]
//	if ok {
//		result.count += 1
//		result.totalSize += f.Size()
//		result.list = append(result.list, path)
//	} else {
//		fm.m[key] = newFile(path, f.Size())
//	}
//	fm.Unlock()
//}
//
//// File
//type File struct {
//	size      int64
//	totalSize int64
//	count     int
//	list      []string
//}
//
//func newFile(path string, size int64) *File {
//	f := File{
//		list:      make([]string, 0, 10),
//		size:      size,
//		totalSize: size,
//		count:     1,
//	}
//	f.list = append(f.list, path)
//
//	return &f
//}
//
//type Files []*File
//
//func (f *Files) Len() int {
//	return len(*f)
//}
//func (f *Files) Less(i, j int) bool {
//	return (*f)[i].totalSize < (*f)[j].totalSize
//}
//func (f *Files) Swap(i, j int) {
//	(*f)[i], (*f)[j] = (*f)[j], (*f)[i]
//}
//func (f *Files) Sort() {
//	sort.Sort(sort.Reverse(f))
//}
//
//// Initialize
//func init() {
//	fm = NewFileMap()
//	t1 = time.Now()
//	runtime.GOMAXPROCS(runtime.NumCPU())
//}
//
// Main
//func main() {
	//fmt.pr
//}
//
//	const (
//		Version = "1.0.1710.12502"
//	)
//
//	// Set and check flags
//	fs = flag.NewFlagSet("", flag.ExitOnError)
//	var (
//		searchDir      = fs.String("d", "", "Source directory")
//		countToDisplay = fs.Int("c", 3, "Minimum count")
//		dispVer        = fs.Bool("v", false, "Print version")
//	)
//	fs.Usage = printHelp
//	fs.Parse(os.Args[1:])
//	if *dispVer == true {
//		fmt.Printf("dufind v%s\n", Version)
//		return
//	}
//	if *searchDir == "" {
//		printHelp()
//		return
//	}
//
//	// Read and organize all files
//	wg := new(sync.WaitGroup)
//	var count int64
//	var dispCount int64
//
//	err := filepath.Walk(*searchDir, func(path string, f os.FileInfo, err error) error {
//		if !f.IsDir() {
//			count += 1
//			wg.Add(1)
//
//			go func(name, path string, size int64) {
//				s := strconv.FormatInt(size, 10)
//				checksum := md5.Sum([]byte(name + s))
//				key := hex.EncodeToString(checksum[:16])
//				fm.Store(key, path, f)
//				wg.Done()
//			}(f.Name(), path, f.Size())
//		}
//
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	wg.Wait()
//
//	// Sort
//	values := make(Files, 0, len(fm.m))
//	for _, value := range fm.m {
//		values = append(values, value)
//	}
//	values.Sort()
//
//	// Print
//	for _, f := range values {
//		if f.count >= *countToDisplay && f.totalSize > 0 {
//			fmt.Printf("# total %-15d bytes, each %-15d bytes, %d files\n", f.totalSize, f.size, f.count)
//			for _, fn := range f.list {
//				fmt.Printf("\t%s\n", fn)
//			}
//			dispCount += int64(f.count)
//		}
//
//	}
//	fmt.Printf("\n# Time: %s, Count(displayed/total): %d/%d, CPU: %d\n", time.Since(t1), dispCount, count, runtime.NumCPU())
//}

func printHelp() {
	fmt.Println("backup [options]")
	fs.PrintDefaults()
}
