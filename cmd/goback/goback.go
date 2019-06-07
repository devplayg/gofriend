package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"

	"github.com/devplayg/yuna/goback"
)

const (
	ProductName = "backup"
	Version     = "1.0.1804.12901"
)

var (
	fs *flag.FlagSet
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		DisableColors: true,
	})
}

func main() {

	// Set CPU count
	runtime.GOMAXPROCS(runtime.NumCPU())

	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		srcDir  = fs.String("s", "", "Source directory")
		dstDir  = fs.String("d", "", "Destination directory")
		version = fs.Bool("v", false, "Version")
		debug   = fs.Bool("debug", false, "Debug")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	if *version {
		fmt.Printf("%s %s\n", ProductName, Version)
		return
	}

	// Check directory parameters
	if *srcDir == "" || *dstDir == "" {
		printHelp()
		return
	}

	// Check source directory
	fi, err := os.Lstat(*srcDir)
	if err != nil {
		log.Error(err)
		return
	}
	if !fi.Mode().IsDir() {
		log.Errorf("invalid source directory: %s", fi.Name())
		return
	}

	// Check destination directory
	fi, err = os.Lstat(*srcDir)
	if err != nil {
		log.Error(err)
		return
	}
	if !fi.Mode().IsDir() {
		log.Errorf("invalid destination directory: %s", fi.Name())
		return
	}

	//	Start backup files
	b := goback.NewBackup(*srcDir, *dstDir, *debug)
	defer b.Close()

	// Initialize backup
	if err := b.Initialize(); err != nil {
		log.Error(err)
		return
	}

	// Start backup
	if err = b.Start(); err != nil {
		log.Error(err)
	}

}

func printHelp() {
	fmt.Println("backup - Backup changed files")
	fmt.Println("backup [options]")
	fmt.Println("ex) backup -s /home/data -d /backup")
	fs.PrintDefaults()
}
