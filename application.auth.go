package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"

	"devops-console/services"
)

type ApplicationAuth struct {
	*Server
}

func (c *Server) Login(ctx iris.Context) {
	s := c.recreateSession(ctx, func(cookie *http.Cookie) {
		// need lax mode for oauth redirect
		cookie.SameSite = http.SameSiteLaxMode
	})

	randReader := rand.Reader
	b := make([]byte, 16)
	if _, err := randReader.Read(b); err != nil {
		c.logger.Error(err)
		c.respondError(ctx, errors.New("unable to start oauth"))
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
	ctx.ViewData("messageSuccess", "Logged out")
	c.templateLogin(ctx, true)
}

func (c *Server) LogoutForced(ctx iris.Context) {
	ctx.ViewData("ERROR_MESSAGE", "Session was terminated, please login again.")
	c.templateLogin(ctx, true)
}

func (c *Server) LoginViaOauth(ctx iris.Context) {
	s := c.startSession(ctx)
	oauth := c.newServiceOauth(ctx)

	if err := ctx.URLParam("error"); err != "" {
		message := err
		if errorDesc := ctx.URLParam("error_description"); errorDesc != "" {
			message = fmt.Sprintf("%s:\n%s", err, errorDesc)
		}
		c.logger.Error(err)
		ctx.ViewData("messageError", message)
		c.templateLogin(ctx, true)
		return
	}

	code := ctx.URLParam("code")
	if code == "" {
		c.destroySession(ctx)
		ctx.ViewData("messageError", "OAuth pre check failed: code empty")
		c.templateLogin(ctx, true)
		return
	}

	state := ctx.URLParam("state")
	if state == "" {
		ctx.ViewData("messageError", "OAuth pre check failed: state empty")
		c.templateLogin(ctx, true)
		return
	}

	if state != s.Get("oauth") {
		ctx.ViewData("messageError", "OAuth pre check failed: state mismatch")
		c.templateLogin(ctx, true)
		return
	}

	tkn, err := oauth.Exchange(code)
	if err != nil {
		c.logger.Error(err)
		ctx.ViewData("messageError", "OAuth check failed: failed getting token from provider")
		c.templateLogin(ctx, true)
		return
	}

	if !tkn.Valid() {
		ctx.ViewData("messageError", "OAuth check failed: invalid token")
		c.templateLogin(ctx, true)
		return
	}

	user, err := oauth.FetchUserInfo(tkn)
	if err != nil {
		c.logger.Error(err)
		ctx.ViewData("messageError", "OAuth check failed: unable to get user information")
		c.templateLogin(ctx, true)
		return
	}

	// check username
	if user.Username == "" {
		ctx.ViewData("messageError", "OAuth login failed: provided username is empty")
		c.templateLogin(ctx, true)
		return
	}

	if c.config.App.Oauth.Filter.UsernameWhitelist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameWhitelist)
		if !filterRegexp.MatchString(user.Username) {
			ctx.ViewData("messageError", fmt.Sprintf("user %s is not allowed to use this application", user.Username))
			c.templateLogin(ctx, true)
			return
		}
	}

	if c.config.App.Oauth.Filter.UsernameBlacklist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameBlacklist)
		if filterRegexp.MatchString(c.config.App.Oauth.Filter.UsernameBlacklist) {
			ctx.ViewData("messageError", fmt.Sprintf("user %s is not allowed to use this application", user.Username))
			c.templateLogin(ctx, true)
			return
		}
	}

	if userSession, err := user.ToJson(); err == nil {
		// regenerate session
		s := c.recreateSession(ctx)

		// inject user information into session
		s.Set("user", userSession)
		c.csrfProtectionTokenRegenerate(ctx)
	} else {
		ctx.ViewData("messageError", "unable to set session")
		c.templateLogin(ctx, true)
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "login"}).Inc()

	c.template(ctx, "Home", "home.jet")
}

func (c *Server) newServiceOauth(ctx iris.Context) services.OAuth {
	oauth := services.OAuth{Host: ctx.Host()}
	oauth.Config = c.config.App.Oauth
	return oauth
}
