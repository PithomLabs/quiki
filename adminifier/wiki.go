package adminifier

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
)

var javascriptTemplates string

var frameHandlers = map[string]func(string, *webserver.WikiInfo, *http.Request) interface{}{
	"dashboard":  handleDashboardFrame,
	"pages":      handlePagesFrame,
	"categories": handleCategoriesFrame,
	"images":     handleImagesFrame,
	"models":     handleModelsFrame,
	"settings":   handleSettingsFrame,
	"edit-page":  handleEditPageFrame,
}

// wikiTemplate members are available to all wiki templates
type wikiTemplate struct {
	User      authenticator.User // user
	Shortcode string             // wiki shortcode
	WikiTitle string             // wiki title
	Static    string             // adminifier-static root
	AdminRoot string             // adminifier root
	Root      string             // wiki root
}

// TODO: verify session on ALL wiki handlers

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {

	// each of these URLs generates wiki.tpl
	for _, which := range []string{
		"dashboard", "pages", "categories",
		"images", "models", "settings", "help", "edit-page",
	} {
		mux.HandleFunc(host+root+shortcode+"/"+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, wi, w, r)
		})
	}

	// frames to load via ajax
	frameRoot := root + shortcode + "/frame/"
	log.Println(frameRoot)
	mux.HandleFunc(host+frameRoot, func(w http.ResponseWriter, r *http.Request) {

		// check logged in
		if !sessMgr.GetBool(r.Context(), "loggedIn") {
			http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
			return
		}

		frameName := strings.TrimPrefix(r.URL.Path, frameRoot)
		tmplName := "frame-" + frameName + ".tpl"

		// frame template does not exist
		if exist := tmpl.Lookup(tmplName); exist == nil {
			http.NotFound(w, r)
			return
		}

		// call func to create template params
		var dot interface{} = nil
		if handler, exist := frameHandlers[frameName]; exist {
			dot = handler(shortcode, wi, r)
		}

		// execute frame template
		err := tmpl.ExecuteTemplate(w, tmplName, dot)

		// error occurred in template execution
		if err != nil {
			panic(err)
		}
	})
}

func handleWiki(shortcode string, wi *webserver.WikiInfo, w http.ResponseWriter, r *http.Request) {

	// check logged in
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
		return
	}

	// load javascript templates
	if javascriptTemplates == "" {
		files, _ := filepath.Glob(dirAdminifier + "/template/js-tmpl/*.tpl")
		for _, fileName := range files {
			data, _ := ioutil.ReadFile(fileName)
			javascriptTemplates += string(data)
		}
	}

	err := tmpl.ExecuteTemplate(w, "wiki.tpl", struct {
		JSTemplates template.HTML
		wikiTemplate
	}{
		template.HTML(javascriptTemplates),
		getWikiTemplate(shortcode, wi, r),
	})
	if err != nil {
		panic(err)
	}
}

func handleDashboardFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	return nil
}

func handlePagesFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	return handleFileFrames(shortcode, wi, r, wi.Pages())
}

func handleImagesFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	return handleFileFrames(shortcode, wi, r, wi.Images(), "d")
}

func handleModelsFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	return handleFileFrames(shortcode, wi, r, wi.Models())
}

func handleCategoriesFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	cats := wi.Categories()
	return handleFileFrames(shortcode, wi, r, cats)
}

func handleFileFrames(shortcode string, wi *webserver.WikiInfo, r *http.Request, results interface{}, extras ...string) interface{} {
	res, err := json.Marshal(map[string]interface{}{
		"sort_types": append([]string{"a", "c", "u", "m"}, extras...),
		"results":    results,
	})
	if err != nil {
		panic(err)
	}
	return struct {
		JSON  template.HTML
		Order string
		wikiTemplate
	}{
		JSON:         template.HTML("<!--JSON\n" + string(res) + "\n-->"),
		Order:        "m-",
		wikiTemplate: getWikiTemplate(shortcode, wi, r),
	}
}

func handleSettingsFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	return nil
}

func handleEditPageFrame(shortcode string, wi *webserver.WikiInfo, r *http.Request) interface{} {
	// TODO: we need a way to present these errors
	q := r.URL.Query()

	// no page filename provided
	name := q.Get("page")
	if name == "" {
		return nil
	}

	// find the page. if File is empty, it doesn't exist
	info := wi.PageInfo(name)
	if info.File == "" {
		return nil
	}

	// call DisplayFile to get the content
	res := wi.DisplayFile(info.Path)
	fileRes, ok := res.(wiki.DisplayFile)
	if !ok {
		return nil
	}

	return struct {
		Found   bool
		Model   bool   // true if editing a model
		Title   string // page title or filename
		File    string // filename
		Content string // file content
		wikiTemplate
	}{
		Found:        true,
		Model:        false,
		Title:        info.Title,
		File:         info.File,
		Content:      fileRes.Content,
		wikiTemplate: getWikiTemplate(shortcode, wi, r),
	}
}

func getWikiTemplate(shortcode string, wi *webserver.WikiInfo, r *http.Request) wikiTemplate {
	return wikiTemplate{
		User:      sessMgr.Get(r.Context(), "user").(authenticator.User),
		Shortcode: shortcode,
		WikiTitle: wi.Title,
		AdminRoot: strings.TrimRight(root, "/"),
		Static:    root + "adminifier-static",
		Root:      root + shortcode,
	}
}