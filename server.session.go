package main

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12/sessions"
)

func (c *Server) initSession() {
	c.logger.Infof(" - using %v session", c.config.App.Session.Type)
	c.logger.Infof("   - cookie name: %v", c.config.App.Session.CookieName)
	c.logger.Infof("   - session expiry: %v", c.config.App.Session.Expiry)

	switch c.config.App.Session.Type {
	case "internal":
		c.initSessionInternal()
		break
	case "securecookie":
		c.initSessionSecureCookie()
		break
	default:
		panic(fmt.Sprintf("Invalid session type defined"))
	}

	c.app.Use(c.session.Handler())
}

func (c *Server) initSessionInternal() {
	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: false,
	})
}

func (c *Server) initSessionSecureCookie() {
	secureCookie := securecookie.New(
		[]byte(c.config.App.Session.SecureCookie.HashKey),
		[]byte(c.config.App.Session.SecureCookie.BlockKey),
	)

	c.session = sessions.New(sessions.Config{
		Cookie:       c.config.App.Session.CookieName,
		Encode:       secureCookie.Encode,
		Decode:       secureCookie.Decode,
		AllowReclaim: true,
		Expires:      c.config.App.Session.Expiry,
	})
}
