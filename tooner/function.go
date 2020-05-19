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

func getContent(list []string, outputFile string) (string, string) {
	var content string
	var folder string

	for _, name := range list {
		matched, _ := regexp.MatchString(`\.(jpg|jpeg|gif|bmp|png)$`, strings.ToLower(name))

		if matched { // is image
			content += fmt.Sprintf("<img src='%s'>", name)
			continue
		}

		folder += fmt.Sprintf("<li><a href='%s/%s'>%s</a></li>", name, outputFile, name)
	}
	if len(folder) > 0 {
		folder = "<ul>" + folder + "</ul>"
	}
	return folder, content
}

func getNavigation(rootDir, dir, prev, next string) (string, string) {
	parentDir := getParentDir(rootDir)

	if rootDir == dir {
		return filepath.Base(rootDir), ""
	}

	str := strings.TrimPrefix(dir, parentDir+string(os.PathSeparator))
	arr := strings.Split(str, string(os.PathSeparator))
	var nav = arr[len(arr)-1]
	for i := len(arr) - 2; i >= 0; i-- {
		path := strings.Repeat("../", len(arr)-1-i)
		nav = fmt.Sprintf("<a href='%sindex.html'>%s</a> &gt; ", path, arr[i]) + nav
	}

	return nav, prev + next
}

func getParentDir(dir string) string {
	parentDir, _ := filepath.Abs(filepath.Join(dir, ".."))
	return parentDir
}

func wrapHtml(folders, content, position, nav string) string {
	html := `<!DOCTYPE html>
<html lang="en-US">
<head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<style>
body{color: #555555;}
.position{display:block; }
.nav{font-size:1.3rem; text-align:center; }
.content{text-align:center; display:block; }
.images{line-height:0px;  display: inline-block; }
img{display:block; max-width: 100%%; height:auto;}
ul{list-style-type: none;  flex-wrap: wrap;
  display: flex;}
ul li {
  flex: 1 0 10%%;
}
a:link{color:#0366d6; text-decoration:none}
a:visited{color:#0366d6;}
a:hover{color: #0366d6; text-decoration:underline;}
a:active{color: #0366d6;}
.next{margin-left:7px;}
</style>
</head>
<body>
<div class="position">%s</div>
<div class="nav">%s</div>
<div class="folders">%s</div>
<div class="content"><div class="images">%s</div></div>
<div class="nav">%s</div>
<div class="position">%s</div>
</body>
</html>
`
	html = fmt.Sprintf(html, position, nav, folders, content, nav, position)
	html = strings.ReplaceAll(html, "\n", "")
	return html
}

func getPrevNextDir(rootDir string, dirs []string, idx int) (string, string) {
	prevIdx := idx - 1
	nextIdx := idx + 1
	if nextIdx > len(dirs)-1 {
		nextIdx = -1
	}

	var prev, next string
	if prevIdx > 0 {
		prev = dirs[prevIdx]
	}
	if nextIdx > 0 {
		next = dirs[nextIdx]
	}

	parentDir := getParentDir(dirs[idx])
	if getParentDir(prev) != parentDir {
		prev = ""
	}
	if getParentDir(next) != parentDir {
		next = ""
	}

	if len(prev) > 0 {
		prevBase := filepath.Base(prev)
		prev = fmt.Sprintf("<a href='../%s/index.html'>%s</a>", prevBase, "Prev")
	}

	if len(next) > 0 {
		nextBase := filepath.Base(next)
		next = fmt.Sprintf("<a href='../%s/index.html' class='next'>%s</a>", nextBase, "Next")
	}
	return prev, next
}
