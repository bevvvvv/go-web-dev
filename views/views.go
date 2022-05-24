package views

import (
	"bytes"
	"go-web-dev/context"
	"html/template"
	"io"
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
	thisView.Render(w, r, nil)
}

func (thisView *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var viewData Data
	switch d := data.(type) {
	case Data:
		viewData = d
	default:
		viewData = Data{
			Yield: d,
		}
	}
	viewData.User = context.User(r.Context())

	var buffer bytes.Buffer
	if err := thisView.Template.ExecuteTemplate(&buffer, thisView.Layout, viewData); err != nil {
		http.Error(w, "Something went wrong. If the problem persists please contact us.", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buffer)
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
