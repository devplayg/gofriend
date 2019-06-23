package main

import (
	"fmt"
	"github.com/devplayg/yuna/tooner"
	"github.com/spf13/pflag"
	"os"
)

const (
	appName    = "Tooner"
	appVersion = "1.0.5"
)

var targetDir string

func main() {
	indexFileName := "index.html"
	tn := tooner.NewTooner(targetDir, indexFileName)
	tn.Start()
}

func init() {
	fs := pflag.NewFlagSet("tooner", pflag.ContinueOnError)

	// Get arguments
	dir := fs.StringP("dir", "d", "", "Directory")
	fs.Usage = func() {
		fmt.Printf("%s v%s\n", appName, appVersion)
		fs.PrintDefaults()
	}
	_ = fs.Parse(os.Args[1:])

	if len(*dir) < 1 {
		os.Exit(1)
	}
	targetDir = *dir
}
