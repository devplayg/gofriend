package main

import (
	"fmt"
	"github.com/devplayg/yuna/dff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"runtime"
)

var fs *pflag.FlagSet

func main() {
	fs = pflag.NewFlagSet("dff", pflag.ContinueOnError)

	dirs := fs.StringArray("dir", []string{}, "target directories to search duplicate files")
	cpu := fs.Int("cpu", 0, "CPU Count to use")
	minFileCount := fs.Int("min-count", 3, "Minimum file count to find")
	minFileSize := fs.Int64("min-size", 100, "Minimum file size to find")
	debug := fs.Bool("debug", false, "debug")

	fs.Usage = printHelp
	_ = fs.Parse(os.Args[1:])

	if *debug {
		log.Info("debug")
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
	fmt.Println("dff - Duplicate file finder")
	fmt.Println("dff [options]")
	fmt.Println("ex) backup -s /home/data -d /backup")
	fs.PrintDefaults()
}
