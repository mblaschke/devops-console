package main

import (
	"devops-console/models"
	"fmt"
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
		if err := ctx.View(template); err != nil {
			c.logger.Errorln(err)
		}
	})
}

func (c *Server) react(ctx iris.Context, title string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context, user *models.User) {
		ctx.ViewData("title", title)
		if err := ctx.View("react.jet"); err != nil {
			c.logger.Errorln(err)
		}
	})
}

func (c *Server) heartbeat(ctx iris.Context) {
	user, err := c.getUser(ctx)
	if err == nil && user != nil {
		ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
		ctx.JSON("Ok")
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.JSON("Failed")
	}
}
