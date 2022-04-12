package controllers

import (
	"go-web-dev/views"
)

func NewStaticController() *StaticController {
	return &StaticController{
		HomeView:    views.NewView("bootstrap", "views/static/home.gohtml"),
		ContactView: views.NewView("bootstrap", "views/static/contact.gohtml"),
	}
}

type StaticController struct {
	HomeView    *views.View
	ContactView *views.View
}
