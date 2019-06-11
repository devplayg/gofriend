package main

import (
	"github.com/devplayg/yuna/dff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"runtime"
)

var fs *pflag.FlagSet

func main() {
	fs = pflag.NewFlagSet("dff", pflag.ContinueOnError)

	// Handle options
	dirs := fs.StringArrayP("dir", "d", []string{}, "target directories to search duplicate files")
	cpu := fs.Int("cpu", 0, "CPU Count to use")
	minFileCount := fs.IntP("min-count", "c", 3, "Minimum file count to find")
	minFileSize := fs.Int64P("min-size", "s", 100, "Minimum file size to find")
	verbose := fs.BoolP("verbose", "v", false, "Verbose")

	fs.Usage = printHelp
	_ = fs.Parse(os.Args[1:])

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	if *cpu == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	duplicateFileFinder := dff.NewDuplicateFileFinder(*dirs, *minFileCount, *minFileSize)
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
