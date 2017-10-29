package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/devplayg/gofriend/backup"
	"github.com/devplayg/gofriend"
	"log"
	"time"
)

var (
	fs *flag.FlagSet
	t1 time.Time
)

const (
	DefaultDBFile = "/home/backup/backup.db"
)

func main() {
	t1 := time.Now()
	fs = flag.NewFlagSet("", flag.ExitOnError)

	var (
		srcDir = fs.String("s", "/home/current/", "Source directory")
		dstDir = fs.String("d", "/home/backup/", "Destination directory")
		db     = fs.String("db", DefaultDBFile, "History file")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	//	backup
	b := backup.NewBackup(*srcDir, *dstDir, *db)
	err := b.Initialize()
	if err != nil {
		log.Println(err)
		return
	}
	err = b.Start()
	gofriend.CheckErr(err)

	log.Printf("Time: %3.1f\n", time.Since(t1).Seconds())
}

func printHelp() {
	fmt.Println("backup [options]")
	fs.PrintDefaults()
}
