package controllers

import (
	"go-web-dev/views"
)

func NewStaticController() *StaticController {
	return &StaticController{
		HomeView:    views.NewView("bootstrap", "static/home"),
		ContactView: views.NewView("bootstrap", "static/contact"),
	}
}

type StaticController struct {
	HomeView    *views.View
	ContactView *views.View
}
