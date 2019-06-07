package main

import (
	"fmt"
	"github.com/devplayg/yuna/dff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
)

var fs *pflag.FlagSet

//
//// -------------------------------------------------
//type arrayFlags []string
//
//func (i *arrayFlags) String() string {
//    return i
//}
//
//func (i *arrayFlags) Set(value string) error {
//    *i = append(*i, value)
//    return nil
//}
//// -------------------------------------------------

//var myFlags arrayFlags
//a -d adsfasdf -d asdfasdf asd -d asdfasdfasdf

func main() {
	fs = pflag.NewFlagSet("dff", pflag.ContinueOnError)
	dirs := fs.StringArray("dir", []string{}, "directory")
	debug := fs.Bool("debug", false, "debug")
	_ = fs.Int("cpu", 1, "CPU Count to use")
	_ = fs.Uint64("min-size", 100000, "Min size")

	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	if *debug {
		log.Info("debug")
		log.SetLevel(log.DebugLevel)
	}

	//spew.Dump(dirs)
	//
	err := dff.Start(*dirs)
	if err != nil {
		log.Error(err)
		return
	}
}

func init() {
	//log.SetFormatter(&log.TextFormatter{
	//    DisableColors: true,
	//    FullTimestamp: true,
	//})
}

//func IsValidDirectory(dir string) error {
//if _, err := os.Stat(dir); os.IsNotExist(err) {
//    return errors.New("directory not found:"+dir)
//}
//
//_, err := os.Stat(dir)
//if err != nil {
//    return err
//}
//
////if !fi.IsDir()
//
//return nil
//}

func printHelp() {
	fmt.Println("dff - Duplicate file finder")
	fmt.Println("dff [options]")
	fmt.Println("ex) backup -s /home/data -d /backup")
	fs.PrintDefaults()
}
