// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"errors"
	wikiclient "github.com/cooper/go-wikiclient"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var templateDirs string
var templates = make(map[string]wikiTemplate)

type wikiTemplate struct {
	path       string             // template directory path
	template   *template.Template // master HTML template
	staticPath string             // static file directory path, if any
	staticRoot string             // static file directory HTTP root, if any
	logo       string             // path for logo file, if any
}

func getTemplate(name string) (wikiTemplate, error) {

	// template is already cached
	if t, ok := templates[name]; ok {
		return t, nil
	}

	for _, templateDir := range strings.Split(templateDirs, ",") {
		templatePath := templateDir + "/" + name
		t, err := loadTemplate(name, templatePath)

		// an error occurred in loading the template
		if err != nil {
			return t, err
		}

		// no template but no error means try the next directory
		if t.template == nil {
			continue
		}

		return t, nil
	}

	// never found a template
	return wikiTemplate{}, errors.New("unable to find template " + name)
}

func loadTemplate(name, templatePath string) (wikiTemplate, error) {
	var t wikiTemplate
	var tryNextDirectory bool

	// parse HTML templates
	tmpl := template.New("")
	err := filepath.Walk(templatePath, func(filePath string, info os.FileInfo, err error) error {

		// walk error, probably missing template
		if err != nil {
			tryNextDirectory = true
			return err
		}

		// found template file
		if strings.HasSuffix(filePath, ".tpl") {

			// error in parsing
			if _, err := tmpl.ParseFiles(filePath); err != nil {
				return err
			}
		}

		// found static content directory
		if info.IsDir() && info.Name() == "static" {
			t.staticPath = filePath
			t.staticRoot = "/tmpl/" + name
			fileServer := http.FileServer(http.Dir(filePath))
			pfx := t.staticRoot + "/"
			http.Handle(pfx, http.StripPrefix(pfx, fileServer))
			log.Printf("[%s] template registered: %s", name, pfx)
		}

		// found logo
		if t.staticRoot != "" && strings.HasPrefix(filePath, t.staticPath+"/logo.") {
			t.logo = t.staticRoot + "/" + info.Name()
		}

		return err
	})

	// not found
	if tryNextDirectory {
		return t, nil
	}

	// other error
	if err != nil {
		return t, err
	}

	// cache the template
	t.path = templatePath
	t.template = tmpl
	templates[name] = t

	return t, nil
}

type wikiPage struct {
	WholeTitle string             // optional, shown in <title> as-is
	Title      string             // page title
	WikiTitle  string             // wiki title
	WikiLogo   string             // path to wiki logo image
	WikiRoot   string             // wiki HTTP root
	Res        wikiclient.Message // response
	StaticRoot string             // path to static resources
	navigation []interface{}      // slice of nav items [display, url]
}

type navItem struct {
	Display string
	Link    string
}

func (p wikiPage) VisibleTitle() string {
	if p.WholeTitle != "" {
		return p.WholeTitle
	}
	return p.Title + " - " + p.WikiTitle
}

func (p wikiPage) PageCSS() template.CSS {
	return template.CSS(p.Res.Get("css"))
}

func (p wikiPage) HTMLContent() template.HTML {
	return template.HTML(p.Res.Get("content"))
}

func (p wikiPage) Navigation() []navItem {

	// no navigation
	if len(p.navigation) != 2 {
		return nil
	}

	// first item is ordered keys, second is values
	displays := p.navigation[0].([]interface{})
	urls := p.navigation[1].([]interface{})
	items := make([]navItem, len(displays))
	for i := 0; i < len(items); i++ {
		items[i] = navItem{
			displays[i].(string),
			urls[i].(string),
		}
	}

	return items
}
