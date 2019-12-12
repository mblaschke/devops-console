package main

import (
	"devops-console/services"
	"encoding/base64"
	"fmt"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"crypto/rand"
)

type ApplicationAuth struct {
	*Server
}

func (c *Server) Login(ctx iris.Context) {
	randReader := rand.Reader
	b := make([]byte, 16)
	randReader.Read(b)

	state := base64.URLEncoding.EncodeToString(b)

	s := c.session.Start(ctx)
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
	ctx.View("login.jet")
}

func (c *Server) LogoutForced(ctx iris.Context) {
	s := c.session.Start(ctx)
	s.Clear()
	s.Destroy()

	ctx.ViewData("ERROR_MESSAGE", "Session was terminated, please login again.")
	ctx.View("login.jet")
}

func (c *Server) LoginViaOauth(ctx iris.Context) {
	s := c.session.Start(ctx)
	oauth := c.newServiceOauth(ctx)

	if error := ctx.URLParam("error"); error != "" {
		message := error
		if errorDesc := ctx.URLParam("error_description"); errorDesc != "" {
			message = fmt.Sprintf("%s:\n%s", error, errorDesc)
		}
		ctx.ViewData("messageError", message)
		ctx.View("login.jet")
		return
	}

	code := ctx.URLParam("code")
	if code == "" {
		ctx.ViewData("messageError", "OAuth pre check failed: code empty")
		ctx.View("login.jet")
		return
	}

	state := ctx.URLParam("state")
	if state == "" {
		ctx.ViewData("messageError", "OAuth pre check failed: state empty")
		ctx.View("login.jet")
		return
	}

	if state != s.Get("oauth") {
		ctx.ViewData("messageError", "OAuth pre check failed: state mismatch")
		ctx.View("login.jet")
		return
	}

	tkn, err := oauth.Exchange(code)
	if err != nil {
		ctx.ViewData("messageError", "OAuth check failed: failed getting token from provider")
		ctx.View("login.jet")
		return
	}

	if !tkn.Valid() {
		ctx.ViewData("messageError", "OAuth check failed: invalid token")
		ctx.View("login.jet")
		return
	}

	user, err := oauth.FetchUserInfo(tkn)
	if err != nil {
		ctx.ViewData("messageError", "OAuth check failed: unable to get user information")
		ctx.View("login.jet")
	}

	// check username
	if user.Username == "" {
		ctx.ViewData("messageError", "OAuth login failed: provided username is empty")
		ctx.View("login.jet")
	}

	if c.config.App.Oauth.Filter.UsernameWhitelist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameWhitelist)
		if !filterRegexp.MatchString(user.Username) {
			ctx.ViewData("messageError", fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			ctx.View("login.jet")
			return
		}
	}

	if c.config.App.Oauth.Filter.UsernameBlacklist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameBlacklist)
		if filterRegexp.MatchString(c.config.App.Oauth.Filter.UsernameBlacklist) {
			ctx.ViewData("messageError", fmt.Sprintf("User %s is not allowed to use this application", user.Username))
			ctx.View("login.jet")
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
		ctx.View("login.jet")
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
