package main

import (
	"github.com/devplayg/yuna/dff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
)

var fs *pflag.FlagSet

func main() {
	fs = pflag.NewFlagSet("dff", pflag.ContinueOnError)

	// Handle options
	dirs := fs.StringArrayP("dir", "d", []string{}, "target directories to search duplicate files")
	cpu := fs.Int("cpu", 0, "CPU count")
	minNumOfFilesInFileGroup := fs.IntP("min-count", "c", 5, "Minimum number of files in file group")
	minFileSize := fs.Int64P("min-size", "s", 1024*10e3, "Minimum file size (Byte)")
	verbose := fs.BoolP("verbose", "v", false, "Verbose")
	sortBy := fs.StringP("sort", "r", "total", "Sort by [size|total|count]")

	fs.Usage = printHelp
	_ = fs.Parse(os.Args[1:])

	duplicateFileFinder := dff.NewDuplicateFileFinder(*dirs, *minNumOfFilesInFileGroup, *minFileSize, *sortBy)
	duplicateFileFinder.Init(*verbose, *cpu)
	err := duplicateFileFinder.Start()
	if err != nil {
		log.Error(err)
		return
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func printHelp() {
	println("dff - Duplicate file finder")
	println("dff [options]")
	println("ex) backup -s /home/data -d /backup")
	fs.PrintDefaults()
}
