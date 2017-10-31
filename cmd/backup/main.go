package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/devplayg/gofriend"
	"github.com/devplayg/gofriend/backup"
)

var (
	fs *flag.FlagSet
	t1 time.Time
)

//const (
//	DefaultDBFile = "/home/backup/backup.db"
//)

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

	err = b.Start()
	gofriend.CheckErr(err)
	log.Println("###")

	log.Printf("Total time: %3.1fs\n", time.Since(t1).Seconds())
}

func printHelp() {
	fmt.Println("backup [options]")
	fs.PrintDefaults()
}
