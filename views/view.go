package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"

	"github.com/wmolicki/bookler/context"
	"github.com/wmolicki/bookler/models"
)

var (
	LayoutDir   string = "templates/layouts/"
	TemplateExt string = ".gohtml"
)

type Data struct {
	User    *models.User
	Message *Message
	Stuff   interface{}
}

func NewView(layout string, files ...string) *View {
	files = append(files, layoutFiles()...)

	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField is not implemented")
		},
		"signedIn": func() bool {
			return false
		},
	}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{Template: t, Layout: layout}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var buf bytes.Buffer
	csrfField := csrf.TemplateField(r)
	user := context.User(r.Context())
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
		"signedIn": func() bool {
			return user != nil
		},
	})

	vd := Data{User: user, Stuff: data}

	if message := GetMessage(w, r); message != nil {
		vd.Message = message
		ClearCookies(w)
	}

	if err := tpl.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}
