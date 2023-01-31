package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

func (c *Server) getServiceConnectionUser(ctx iris.Context) (user *models.User) {
	authToken := ctx.GetHeader("Authorization")

	if authToken == "" {
		c.respondError(ctx, fmt.Errorf("missing Authorization header"))
		ctx.StopExecution()
		panic(ctx)
	}

	if !strings.HasPrefix(authToken, "Bearer ") {
		c.respondError(ctx, fmt.Errorf("wrong or invalid Authorization method"))
		ctx.StopExecution()
		panic(ctx)
	}

	authToken = strings.TrimPrefix(authToken, "Bearer ")
	if authToken == "" {
		c.respondError(ctx, fmt.Errorf("empty Authorization token"))
		ctx.StopExecution()
		panic(ctx)
	}

	for teamName, teamConfig := range c.config.Permissions.Team {
		for _, serviceConnction := range teamConfig.ServiceConnections {
			if len(serviceConnction.Token) > 0 && serviceConnction.Token == authToken {
				user = &models.User{}
				user.Uuid = "ServiceConnection"
				user.Username = teamName
				user.Teams = []models.Team{
					{
						Name:                 teamName,
						K8sPermissions:       teamConfig.K8sRoleBinding,
						AzureRoleAssignments: teamConfig.AzureRoleAssignments,
					},
				}
			}
		}
	}

	return
}

func (c *Server) ensureServiceUser(ctx iris.Context, callback func(ctx iris.Context, user *models.User)) {
	user := c.getServiceConnectionUser(ctx)
	if user == nil {
		c.handleError(ctx, errors.New("invalid serviceconnection"), true)
		return
	}

	ctx.ViewData("user", user)
	ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
	callback(ctx, user)
}

func (c *Server) ensureLoggedIn(ctx iris.Context, callback func(ctx iris.Context, user *models.User)) {
	user, err := c.getUser(ctx)

	if err != nil {
		c.handleErrorWithStatus(ctx, http.StatusNotFound, errors.New("invalid session or not logged in"), true)
		return
	}

	ctx.ViewData("user", user)
	ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
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
