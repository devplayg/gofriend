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
	runtime.GOMAXPROCS(runtime.NumCPU())

	t1 := time.Now()
	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		srcDir = fs.String("s", "/home/current/", "Source directory")
		dstDir = fs.String("d", "/home/backup/", "Destination directory")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	//	backup
	b := backup.NewBackup(*srcDir, *dstDir)
	err := b.Initialize()
	if err != nil {
		log.Println(err)
		return
	}

	s, err := b.Start()
	b.Close()
	checkErr(err)

	log.Printf("[Backup] ID=%d, Files: %d (Modified: %d, Added: %d, Deleted: %d), Size: %d, Time: %3.1f\n", s.ID, s.BackupModified+s.BackupAdded+s.BackupDeleted, s.BackupModified, s.BackupAdded, s.BackupDeleted, s.BackupSize, s.BackupTime)
	log.Printf("[Logging] Time: %3.1f\n", s.LoggingTime)
	log.Printf("[Total] Files: %d, Size: %d, Time: %3.1f\n", s.TotalCount, s.TotalSize, time.Since(t1).Seconds())
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
