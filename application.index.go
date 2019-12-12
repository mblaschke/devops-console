package main

import (
	"devops-console/models"
	"github.com/kataras/iris/v12"
)

type ApplicationIndex struct {
	*Server
}

func (c *Server) index(ctx iris.Context) {
	user, err := c.getUser(ctx)

	if err == nil && user != nil {
		ctx.Redirect("/kubernetes/namespaces")
	} else {
		c.templateLogin(ctx)
	}
}

func (c *Server) template(ctx iris.Context, title, template string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		ctx.ViewData("title", title)
		ctx.View(template)
	})
}

func (c *Server) react(ctx iris.Context, title string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		ctx.ViewData("title", title)
		ctx.View("react.jet")
	})
}
