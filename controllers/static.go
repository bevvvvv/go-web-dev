package controllers

import (
	"go-web-dev/views"
)

func NewStaticController() *StaticController {
	return &StaticController{
		HomeView:    views.NewView("bootstrap", "static/home"),
		ContactView: views.NewView("bootstrap", "static/contact"),
		FAQView:     views.NewView("bootstrap", "static/faq"),
	}
}

type StaticController struct {
	HomeView    *views.View
	ContactView *views.View
	FAQView     *views.View
}
