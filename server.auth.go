package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kataras/iris/v12"

	"devops-console/models"
)

func (c *Server) templateLogin(ctx iris.Context) {
	ctx.ViewData("title", "Login")
	if err := ctx.View("login.jet"); err != nil {
		c.logger.Error(err)
	}
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
	c.startSession(ctx)
	user, err := c.getUser(ctx)

	if err != nil {
		c.handleError(ctx, errors.New("invalid session or not logged in"), true)
		return
	}

	ctx.ViewData("user", user)
	ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
	callback(ctx, user)
}

func (c *Server) getUser(ctx iris.Context) (user *models.User, err error) {
	s := c.startSession(ctx)
	userJson := s.GetString("user")

	if opts.Debug && len(userJson) == 0 {
		if val := os.Getenv("DEBUG_SESSION_USER"); val != "" {
			s.Set("user", userJson)
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
		c.handleError(ctx, errors.New("invalid session or not logged in"), true)
	}

	return
}
