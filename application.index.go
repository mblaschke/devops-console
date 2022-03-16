package main

import (
	"fmt"

	"github.com/kataras/iris/v12"

	"devops-console/models"
)

type ApplicationIndex struct {
	*Server
}

func (c *Server) index(ctx iris.Context) {
	user, err := c.getUser(ctx)

	if err == nil && user != nil {
		ctx.Redirect("/kubernetes/namespaces")
	} else {
		c.destroySession(ctx)
		c.templateLogin(ctx)
	}
}

func (c *Server) redirectWithHtml(ctx iris.Context, title, redirect string) {
	ctx.ViewData("title", title)
	ctx.ViewData("redirect", redirect)
	if err := ctx.View("redirect.jet"); err != nil {
		c.logger.Error(err)
	}
}

func (c *Server) template(ctx iris.Context, title, template string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		ctx.ViewData("title", title)
		if err := ctx.View(template); err != nil {
			c.logger.Error(err)
		}
	})
}

func (c *Server) react(ctx iris.Context, title string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		ctx.ViewData("title", title)
		if err := ctx.View("react.jet"); err != nil {
			c.logger.Error(err)
		}
	})
}

func (c *Server) heartbeat(ctx iris.Context) {
	user, err := c.getUser(ctx)
	if err == nil && user != nil {
		ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
		if _, err = ctx.JSON("Ok"); err != nil {
			c.logger.Errorf("error while sending response: %v", err)
		}
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		if _, err := ctx.JSON("Failed"); err != nil {
			c.logger.Errorf("error while sending response: %v", err)
		}
	}
}
