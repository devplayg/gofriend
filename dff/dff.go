package dff

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var fileMap sync.Map

type Finder struct {
	finish chan bool
	dirs   []string
}

func Start(dirs []string) error {
	err := isValidDirectories(dirs)
	if err != nil {
		return err
	}

	//finder := Finder{
	//    finish: make(chan bool),
	//    dirs: dirs,
	//}

	ctx := context.Background()
	waitGroup := new(sync.WaitGroup)

	//go myFunc(&waitgroup)
	//waitgroup.Wait()

	for _, d := range dirs {
		waitGroup.Add(1)
		go collectFiles(ctx, waitGroup, d)
	}
	//<- finder.finish

	waitGroup.Wait()
	log.Debug("collecting finished")

	//if err != nil {
	//    return err
	//}

	//compareAllFiles()

	//print()

	//println(dirs)
	return nil
}

func collectFiles(waitGroup *sync.WaitGroup, dir string) error {
	log.Debugf("collecting directory: %s", dir)
	defer func() {
		log.Debugf("finished collecting directory: %s", dir)
		waitGroup.Done()
	}()
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		//if !f.IsDir() {
		//    fi := newFile(path, f.Size(), f.ModTime())
		//    newMap.Store(path, fi)
		//    b.S.TotalCount += 1
		//    b.S.TotalSize += uint64(f.Size())
		//}
		return nil
	})
	time.Sleep(3 * time.Second)

	return err
}

func findAllFiles(dirs []string) error {
	//ch := make(chan bool)

	return nil
}

func isValidDirectories(dirs []string) error {
	for _, d := range dirs {
		_, err := os.Stat(d)
		if err != nil {
			return err
		}
	}

	return nil
}
