package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/devplayg/gofriend/backup"
	"github.com/dustin/go-humanize"
)

const (
	ProductName = "backup"
	Version     = "1.0.1711.11102"
)

var (
	fs *flag.FlagSet
)

func main() {

	// Set CPU count
	runtime.GOMAXPROCS(runtime.NumCPU())

	//t := time.Now()
	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		srcDir  = fs.String("s", "", "SOURCE")
		dstDir  = fs.String("d", "", "DESTINATION")
		version = fs.Bool("v", false, "Version")
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
	b := backup.NewBackup(*srcDir, *dstDir)
	err := b.Initialize()
	if err != nil {
		log.Println(err)
		return
	}
	err = b.Start()
	checkErr(err)
	b.Close()

	//if s != nil {
	log.Printf("[Backup] ID=%d, Files: %d (Modified: %d, Added: %d, Deleted: %d), Size: %s\n", b.S.ID, b.S.BackupModified+b.S.BackupAdded+b.S.BackupDeleted, b.S.BackupModified, b.S.BackupAdded, b.S.BackupDeleted, humanize.Bytes(b.S.BackupSize))
	log.Printf("[Total] Files: %d, Size: %s\n", b.S.TotalCount, humanize.Bytes(b.S.TotalSize))

	msg := fmt.Sprintf("Total: %3.1fs (Reading: %3.1fs, Comparison: %3.1fs, Logging: %3.1fs)",
		b.S.ExecutionTime,
		b.S.ReadingTime.Sub(b.S.Date).Seconds(),
		b.S.ComparisonTime.Sub(b.S.ReadingTime).Seconds(),
		b.S.LoggingTime.Sub(b.S.ComparisonTime).Seconds(),
	)
	log.Printf("[Time] %s\n", msg)
	//}
}

func printHelp() {
	fmt.Println("backup - Backup changed files")
	fmt.Println("backup [options]")
	fmt.Println("ex) backup -s /home/data -d /backup")
	fs.PrintDefaults()
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
