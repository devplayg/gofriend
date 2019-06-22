package tooner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getFileMap(targetDir, indexFileName string) (FileMap, error) {
	fileMap := make(FileMap)

	err := filepath.Walk(targetDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if f.IsDir() {
			fileMap[path] = make([]string, 0)
			if path != targetDir {
				parentDir, _ := filepath.Abs(filepath.Join(path, ".."))
				fileMap[parentDir] = append(fileMap[parentDir], f.Name())
			}
			return nil
		}

		if f.Name() == indexFileName {
			return nil
		}

		dir := filepath.Dir(path)
		fileMap[dir] = append(fileMap[dir], f.Name())
		return nil
	})
	return fileMap, err
}

func getContent(list []string, indexFileName string) string {
	var content string
	var folder string
	for _, f := range list {
		//matched, err := regexp.Match(`(jpeg|)$`, []byte(`seafood`))
		matched, _ := regexp.MatchString(`\.(jpg|jpeg|gif|bmp|png)$`, strings.ToLower(f))

		if matched { // is image
			content += fmt.Sprintf("<div><img src='%s'></div>", f)
			continue
		}
		folder += fmt.Sprintf("<li><a href='%s/%s'>[%s]</a></li>", f, indexFileName, f)
	}
	if len(folder) > 0 {
		folder = "<ul>" + folder + "</ul>"
	}
	return folder + content
}

func getTitle(rootDir, dir string) string {
	parentDir, _ := filepath.Abs(filepath.Join(rootDir, ".."))
	if rootDir == dir {
		return filepath.Base(rootDir)
	}

	str := strings.TrimPrefix(dir, parentDir+string(os.PathSeparator))
	arr := strings.Split(str, string(os.PathSeparator))
	var title = arr[len(arr)-1]
	for i := len(arr) - 2; i >= 0; i-- {
		path := strings.Repeat("../", len(arr)-1-i)
		title = fmt.Sprintf("<a href='%sindex.html'>%s</a> &gt; ", path, arr[i]) + title
	}
	return title
}

func wrapHtml(content string) string {
	html := `<!DOCTYPE html><html lang="en-US"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<style>
body {olor: #555555; line-height:0px}
a:link{color:#0366d6; text-decoration:none}
a:visited{color:#0366d6;}
a:hover{color: #0366d6; text-decoration:underline}
a:active{color: #0366d6 ;}
</style>
</head>
<body>
`
	html += content + `
</body>
</html>`
	return html
}
