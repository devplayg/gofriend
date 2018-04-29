package main

import (
	"os"
	"runtime"
	"flag"
	"path/filepath"
	"fmt"
)


var (
	fs *flag.FlagSet
)

func main() {
	// Set CPU count
	runtime.GOMAXPROCS(runtime.NumCPU())

	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		dir= fs.String("d", "", "Directory")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	fileExtMap := make(map[string]int)

	filepath.Walk(*dir, func(path string, f os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		fileExtMap[ext]++
		return nil
	})

	for ext, count := range fileExtMap {
		fmt.Printf("%d = %s\n", count, ext)
	}

}

func printHelp() {
	fs.PrintDefaults()
}