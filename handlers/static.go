package handlers

import "github.com/wmolicki/bookler/views"

func NewStatic() *Static {
	return &Static{
		Index: views.NewView("bootstrap", "templates/index.gohtml"),
		About: views.NewView("bootstrap", "templates/about.gohtml"),
	}
}

type Static struct {
	Index *views.View
	About *views.View
}
