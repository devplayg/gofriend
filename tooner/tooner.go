package tooner

import (
	"fmt"
	"github.com/facette/natsort"
	"io/ioutil"
	"path/filepath"
)

type FileMap map[string][]string

func NewTooner(dir, indexFileName string) *WebtoonViewer {
	absDir, _ := filepath.Abs(dir)
	return &WebtoonViewer{
		dir:           absDir,
		indexFileName: indexFileName,
	}
}

type WebtoonViewer struct {
	dir           string
	indexFileName string
}

func (w *WebtoonViewer) Start() error {
	fileMap, err := getFileMap(w.dir, w.indexFileName)
	if err != nil {
		panic(err)
	}

	dirs := make([]string, 0, len(fileMap))
	for key := range fileMap {
		dirs = append(dirs, key)
	}
	natsort.Sort(dirs)

	for idx, dir := range dirs {
		list := fileMap[dir]
		natsort.Sort(list)
		prev, next := getPrevNextDir(w.dir, dirs, idx)
		position, nav := getNavigation(list, w.dir, dir, prev, next)
		folders, content := getContent(list, w.indexFileName)
		outputFile := filepath.Join(dir, w.indexFileName)
		err := ioutil.WriteFile(outputFile, []byte(wrapHtml(folders, content, position, nav)), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s is created.\n", outputFile)
	}

	return nil
}
