package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
		Layout:   layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (thisView *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := thisView.Render(w, nil); err != nil {
		panic(err)
	}
}

func (thisView *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return thisView.Template.ExecuteTemplate(w, thisView.Layout, data)
}

// returns a slice of strings listing all layout files
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(nil)
	}
	return files
}

// Takes in a slice of strings and prepends TemplateDir.
func addTemplatePath(files []string) {
	for ind, file := range files {
		files[ind] = TemplateDir + file
	}
}

// Takes in a slice of strings and appends TemplateExt.
func addTemplateExt(files []string) {
	for ind, file := range files {
		files[ind] = file + TemplateExt
	}
}
