package handlers

import "github.com/wmolicki/bookler/views"

func NewStatic() *Static {
	return &Static{
		Index: views.NewView("bulma", "templates/index.gohtml"),
		About: views.NewView("bulma", "templates/about.gohtml"),
	}
}

type Static struct {
	Index *views.View
	About *views.View
}
