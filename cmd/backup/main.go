package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/devplayg/gofriend/backup"
	"github.com/dustin/go-humanize"
)

const (
	ProductName = "backup"
	Version     = "1.0.1711.11101"
)

var (
	fs *flag.FlagSet
	t1 time.Time
)

func main() {

	// Set CPU count
	runtime.GOMAXPROCS(runtime.NumCPU())

	t := time.Now()
	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		srcDir  = fs.String("s", "", "Source directory")
		dstDir  = fs.String("d", "", "Destination directory")
		dispVer = fs.Bool("v", false, "Version")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	if *dispVer {
		fmt.Printf("%s %s\n", ProductName, Version)
		return
	}

	if *srcDir == "" || *dstDir == "" {
		fs.PrintDefaults()
		return
	}

	//	backup
	backup := backup.NewBackup(*srcDir, *dstDir)
	err := backup.Initialize()
	if err != nil {
		log.Println(err)
		return
	}

	s, err := backup.Start()
	checkErr(err)
	defer backup.Close()

	if s != nil {
		log.Printf("[Backup] ID=%d, Files: %d (Modified: %d, Added: %d, Deleted: %d), Size: %s, Time: %3.1fs\n", s.ID, s.BackupModified+s.BackupAdded+s.BackupDeleted, s.BackupModified, s.BackupAdded, s.BackupDeleted, humanize.Bytes(s.BackupSize), s.BackupTime)
		log.Printf("[Logging] Time: %3.1f\n", s.LoggingTime)
		log.Printf("[Total] Files: %d, Size: %s, Time: %3.1fs\n", s.TotalCount, humanize.Bytes(s.TotalSize), time.Since(t).Seconds())

	}
}

func printHelp() {
	fmt.Println("backup [options]")
	fs.PrintDefaults()
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
