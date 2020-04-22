package main

import (
	"devops-console/models"
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"os"
)

func (c *Server) templateLogin(ctx iris.Context) {
	ctx.ViewData("title", "Login")
	if err := ctx.View("login.jet"); err != nil {
		c.logger.Errorln(err)
	}
}

func (c *Server) ensureLoggedIn(ctx iris.Context, callback func(ctx iris.Context, user *models.User)) {
	c.startSession(ctx)
	user, err := c.getUser(ctx)

	if err != nil {
		c.handleError(ctx, errors.New("Invalid session or not logged in"), true)
		return
	}

	ctx.ViewData("user", user)
	ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
	callback(ctx, user)
}

func (c *Server) getUser(ctx iris.Context) (user *models.User, err error) {
	s := c.startSession(ctx)
	userJson := s.GetString("user")

	if opts.Debug {
		if val := os.Getenv("DEBUG_SESSION_USER"); val != "" {
			s.Set("user", "DEBUG_SESSION_USER")
			userJson = val
		}
	}

	user, err = models.UserCreateFromJson(userJson, &c.config)
	return
}

func (c *Server) getUserOrStop(ctx iris.Context) (user *models.User) {
	var err error
	user, err = c.getUser(ctx)

	if err != nil || user == nil {
		c.handleError(ctx, errors.New("Invalid session or not logged in"), true)
	}

	return
}
