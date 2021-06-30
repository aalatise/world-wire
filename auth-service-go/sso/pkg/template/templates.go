package template

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	TemplateCache map[string]*template.Template
)

func InitTemplateCache(dir string) {
	cache := map[string]*template.Template{}

	cleanRoot := filepath.Clean(dir)
	pfx := len(cleanRoot) + 1

	tmplList := make([]string, 0)

	filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if !info.IsDir() && (strings.HasSuffix(path, ".layout.tmpl") || strings.HasSuffix(path, ".partial.tmpl")) {
			//logging.InfoLogger.Println("Adding template:" + path)
			tmplList = append(tmplList, path)
		}

		return nil
	})

	filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".page.tmpl") {
			root := template.New("")
			//logging.InfoLogger.Println("Adding template to page:" + path)
			if e1 != nil {
				return e1
			}

			b, e2 := ioutil.ReadFile(path)
			if e2 != nil {
				return e2
			}

			name := path[pfx:]
			t := root.New(name)
			t, e2 = t.Parse(string(b))

			if e2 != nil {
				return e2
			}

			t.ParseFiles(tmplList...)

			//logging.InfoLogger.Println("Adding template to template cache map with key: " + name)
			cache[name] = t
		}

		return nil
	})

	TemplateCache = cache
}
