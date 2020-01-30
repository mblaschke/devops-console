package main

import (
	"crypto/rand"
	"devops-console/services"
	"encoding/base64"
	"errors"
	"fmt"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
)

type ApplicationAuth struct {
	*Server
}

func (c *Server) Login(ctx iris.Context) {
	s := c.session.Start(ctx)

	randReader := rand.Reader
	b := make([]byte, 16)
	if _, err := randReader.Read(b); err != nil {
		c.logger.Errorln(err)
		c.respondError(ctx, errors.New("Unable to start oauth"))
		return
	}

	state := base64.URLEncoding.EncodeToString(b)
	s.Set("oauth", state)

	oauth := c.newServiceOauth(ctx)
	url := oauth.AuthCodeURL(state)

	PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "start"}).Inc()

	ctx.Redirect(url)
}

func (c *Server) Logout(ctx iris.Context) {
	s := c.session.Start(ctx)
	s.Clear()
	s.Destroy()

	ctx.ViewData("messageSuccess", "Logged out")
	c.templateLogin(ctx)
}

func (c *Server) LogoutForced(ctx iris.Context) {
	s := c.session.Start(ctx)
	s.Clear()
	s.Destroy()

	ctx.ViewData("ERROR_MESSAGE", "Session was terminated, please login again.")
	c.templateLogin(ctx)
}

func (c *Server) LoginViaOauth(ctx iris.Context) {
	s := c.session.Start(ctx)
	oauth := c.newServiceOauth(ctx)

	if error := ctx.URLParam("error"); error != "" {
		message := error
		if errorDesc := ctx.URLParam("error_description"); errorDesc != "" {
			message = fmt.Sprintf("%s:\n%s", error, errorDesc)
		}
		c.logger.Errorln(error)
		ctx.ViewData("messageError", message)
		c.templateLogin(ctx)
		return
	}

	code := ctx.URLParam("code")
	if code == "" {
		ctx.ViewData("messageError", "OAuth pre check failed: code empty")
		c.templateLogin(ctx)
		return
	}

	state := ctx.URLParam("state")
	if state == "" {
		ctx.ViewData("messageError", "OAuth pre check failed: state empty")
		c.templateLogin(ctx)
		return
	}

	if state != s.Get("oauth") {
		ctx.ViewData("messageError", "OAuth pre check failed: state mismatch")
		c.templateLogin(ctx)
		return
	}

	tkn, err := oauth.Exchange(code)
	if err != nil {
		c.logger.Errorln(err)
		ctx.ViewData("messageError", "OAuth check failed: failed getting token from provider")
		c.templateLogin(ctx)
		return
	}

	if !tkn.Valid() {
		ctx.ViewData("messageError", "OAuth check failed: invalid token")
		c.templateLogin(ctx)
		return
	}

	user, err := oauth.FetchUserInfo(tkn)
	if err != nil {
		c.logger.Errorln(err)
		ctx.ViewData("messageError", "OAuth check failed: unable to get user information")
		c.templateLogin(ctx)
		return
	}

	// check username
	if user.Username == "" {
		ctx.ViewData("messageError", "OAuth login failed: provided username is empty")
		c.templateLogin(ctx)
		return
	}

	if c.config.App.Oauth.Filter.UsernameWhitelist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameWhitelist)
		if !filterRegexp.MatchString(user.Username) {
			ctx.ViewData("messageError", fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			c.templateLogin(ctx)
			return
		}
	}

	if c.config.App.Oauth.Filter.UsernameBlacklist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameBlacklist)
		if filterRegexp.MatchString(c.config.App.Oauth.Filter.UsernameBlacklist) {
			ctx.ViewData("messageError", fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			c.templateLogin(ctx)
			return
		}
	}

	if userSession, err := user.ToJson(); err == nil {
		// regenerate session
		s.Clear()
		s.Destroy()
		s := c.session.Start(ctx)

		s.Set("user", userSession)
		c.csrfProtectionTokenRegenerate(ctx)
	} else {
		ctx.ViewData("messageError", "Unable to set session")
		c.templateLogin(ctx)
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "login"}).Inc()

	ctx.Redirect("/kubernetes/namespaces")
}

func (c *Server) newServiceOauth(ctx iris.Context) services.OAuth {
	oauth := services.OAuth{Host: ctx.Host()}
	oauth.Config = c.config.App.Oauth
	return oauth
}
