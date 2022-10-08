package main

import (
	"fmt"

	iris "github.com/kataras/iris/v12"

	"github.com/mblaschke/devops-console/models"
)

type ApplicationIndex struct {
	*Server
}

func NewApplicationIndex(c *Server) *ApplicationIndex {
	app := ApplicationIndex{Server: c}
	return &app
}

func (c *Server) index(ctx iris.Context) {
	user, err := c.getUser(ctx)

	if err == nil && user != nil {
		ctx.Redirect("/home")
	} else {
		c.templateLogin(ctx, true)
	}
}

func (c *Server) home(ctx iris.Context) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		c.renewSession(ctx)
		c.template(ctx, "Home", "home.jet")
	})
}

func (c *Server) template(ctx iris.Context, title, template string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		ctx.ViewData("title", title)
		if err := ctx.View("pages/" + template); err != nil {
			c.logger.Error(err)
		}
	})
}

func (c *Server) redirectHtml(ctx iris.Context, url string) {
	ctx.ViewData("redirectUrl", url)
	c.template(ctx, "Home", "redirect.jet")
}

func (c *Server) react(ctx iris.Context, title string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		c.renewSession(ctx)
		ctx.ViewData("title", title)
		if err := ctx.View("pages/react.jet"); err != nil {
			c.logger.Error(err)
		}
	})
}

func (c *Server) heartbeat(ctx iris.Context) {
	user, err := c.getUser(ctx)
	if err == nil && user != nil {
		c.renewSession(ctx)
		ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
		c.responseJson(ctx, "Ok")
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		c.responseJson(ctx, "Failed")
	}
}
