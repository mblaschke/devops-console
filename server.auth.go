package main

import (
	"errors"
	"fmt"
	"net/http"

	iris "github.com/kataras/iris/v12"

	"github.com/mblaschke/devops-console/models"
)

func (c *Server) templateLogin(ctx iris.Context, logout bool) {
	if logout {
		c.destroySession(ctx)
	}

	ctx.ViewData("title", "Login")
	ctx.ViewData("redirectUrl", "")
	if err := ctx.View("pages/login.jet"); err != nil {
		c.logger.Error(err)
	}
	ctx.StopExecution()
	panic(ctx)
}

func (c *Server) ensureLoggedIn(ctx iris.Context, callback func(ctx iris.Context, user *models.User)) {
	user, err := c.getUser(ctx)

	if err != nil {
		c.handleErrorWithStatus(ctx, http.StatusNotFound, errors.New("invalid session or not logged in"), true)
		return
	}

	ctx.ViewData("user", user)
	ctx.Values().Set("userIdentification", fmt.Sprintf("upn:%v oid:%v", user.Username, user.Uuid))
	callback(ctx, user)
}

func (c *Server) getUser(ctx iris.Context) (user *models.User, err error) {
	s := c.getSession(ctx)
	if s == nil {
		return nil, fmt.Errorf(`not logged in`)
	}

	userJson := s.GetString("user")
	if len(userJson) >= 1 {
		user, err = models.UserCreateFromJson(userJson, &c.config)
	} else {
		return nil, fmt.Errorf(`not logged in`)
	}
	return
}
