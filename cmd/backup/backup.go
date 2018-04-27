package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	log "github.com/sirupsen/logrus"

	"github.com/devplayg/yuna/backup"
	"github.com/dustin/go-humanize"
)

const (
	ProductName = "backup"
	Version     = "1.0.1804.12702"
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

	//t := time.Now()
	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		srcDir  = fs.String("s", "", "Source directory")
		dstDir  = fs.String("d", "", "Destination directory")
		version = fs.Bool("v", false, "Version")
		debug = fs.Bool("debug", false, "Debug")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	if *version {
		fmt.Printf("%s %s\n", ProductName, Version)
		return
	}

	if *srcDir == "" || *dstDir == "" {
		printHelp()
		return
	}

	//	backup
	b := backup.NewBackup(*srcDir, *dstDir, *debug)
	defer b.Close()
	err := b.Initialize()
	if err != nil {
		log.Error(err)
		return
	}
	if err = b.Start(); err != nil {
		log.Error(err)
	}

	//if s != nil {
	log.Infof("[Backup] ID=%d, Files: %d (Modified: %d, Added: %d, Deleted: %d), Size: %s", b.S.ID, b.S.BackupModified+b.S.BackupAdded+b.S.BackupDeleted, b.S.BackupModified, b.S.BackupAdded, b.S.BackupDeleted, humanize.Bytes(b.S.BackupSize))
	log.Infof("[Total] Files: %d, Size: %s", b.S.TotalCount, humanize.Bytes(b.S.TotalSize))

	msg := fmt.Sprintf("Total: %3.1fs (Reading: %3.1fs, Comparison: %3.1fs, Logging: %3.1fs)",
		b.S.ExecutionTime,
		b.S.ReadingTime.Sub(b.S.Date).Seconds(),
		b.S.ComparisonTime.Sub(b.S.ReadingTime).Seconds(),
		b.S.LoggingTime.Sub(b.S.ComparisonTime).Seconds(),
	)
	log.Debugf("[Time] %s", msg)
}

func printHelp() {
	fmt.Println("backup - Backup changed files")
	fmt.Println("backup [options]")
	fmt.Println("ex) backup -s /home/data -d /backup")
	fs.PrintDefaults()
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
	}
}
