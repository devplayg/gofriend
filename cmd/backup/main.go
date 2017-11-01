package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/devplayg/gofriend/backup"
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
		srcDir = fs.String("s", "c:/", "Source directory")
		dstDir = fs.String("d", "c:/temp", "Destination directory")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	if *srcDir == "" || *dstDir == "" {
		fs.PrintDefaults()
		return
	}

	//	backup
	b := backup.NewBackup(*srcDir, *dstDir)
	err := b.Initialize()
	if err != nil {
		log.Println(err)
		return
	}

	s, err := b.Start()
	checkErr(err)
	b.Close()

	if s != nil {
		log.Printf("[Backup] ID=%d, Files: %d (Modified: %d, Added: %d, Deleted: %d), Size: %d, Time: %3.1fs\n", s.ID, s.BackupModified+s.BackupAdded+s.BackupDeleted, s.BackupModified, s.BackupAdded, s.BackupDeleted, s.BackupSize, s.BackupTime)
		log.Printf("[Logging] Time: %3.1f\n", s.LoggingTime)
		log.Printf("[Total] Files: %d, Size: %d, Time: %3.1fs\n", s.TotalCount, s.TotalSize, time.Since(t).Seconds())
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
