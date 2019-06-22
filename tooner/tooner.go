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

	for dir, list := range fileMap {
		natsort.Sort(list)
		content := getContent(list, w.indexFileName)
		title := getTitle(w.dir, dir)
		nav := fmt.Sprintf("<h3>%s</h3>", title)
		content = nav + "<hr>" + content + "<hr>" + nav
		outputFile := filepath.Join(dir, w.indexFileName)
		err := ioutil.WriteFile(outputFile, []byte(wrapHtml(content)), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s is created.\n", outputFile)
	}
	return nil
}
