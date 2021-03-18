package main

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"math"
	"net/http"
	"time"
)

func (c *Server) startSession(ctx iris.Context) *sessions.Session {
	return c.session.Start(ctx)
}

func (c *Server) initSession() {
	c.logger.Infof(" - using %v session", c.config.App.Session.Type)
	c.logger.Infof("   - cookie name: %v", c.config.App.Session.CookieName)
	c.logger.Infof("   - session expiry: %v", c.config.App.Session.Expiry)

	switch c.config.App.Session.Type {
	case "internal":
		c.initSessionInternal()
	case "securecookie":
		c.initSessionSecureCookie()
	case "redis":
		c.initSessionRedis()
	default:
		panic("invalid session type defined")
	}

	c.app.Use(c.session.Handler(func(cookie *http.Cookie) {
		if c.config.App.Session.CookieSecure {
			cookie.Secure = true
		}

		if c.config.App.Session.CookieDomain != "" {
			cookie.Domain = c.config.App.Session.CookieDomain
		}
	}))
}

func (c *Server) initSessionInternal() {
	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: true,
	})
}

func (c *Server) initSessionSecureCookie() {
	secureCookie := securecookie.New(
		[]byte(c.config.App.Session.SecureCookie.HashKey),
		[]byte(c.config.App.Session.SecureCookie.BlockKey),
	)

	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		Encode:                      secureCookie.Encode,
		Decode:                      secureCookie.Decode,
		AllowReclaim:                true,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: true,
	})
}

func (c *Server) initSessionRedis() {
	for i := 0; i < 25; i++ {
		durationTime := math.Min(15, float64(i*2))
		retryTime := time.Duration(time.Duration(durationTime) * time.Second)

		c.redisConnection = redis.New(redis.Config{
			Network:   "tcp",
			Addr:      c.config.App.Session.Redis.Addr,
			Timeout:   c.config.App.Session.Redis.Timeout,
			MaxActive: c.config.App.Session.Redis.MaxActive,
			Password:  c.config.App.Session.Redis.Password,
			Database:  c.config.App.Session.Redis.Database,
			Prefix:    c.config.App.Session.Redis.Prefix,
			Delim:     c.config.App.Session.Redis.Delim,
			Driver:    redis.Redigo(),
		})

		if c.redisConnection != nil {
			break
		}

		c.logger.Errorln(fmt.Sprintf("redis connection failed, retrying in %v", retryTime.String()))
		time.Sleep(retryTime)
	}

	if c.redisConnection == nil {
		c.logger.Fatalln("redis connection failed, cannot connect to session database")
	}

	c.logger.Infoln("redis connection established")

	// Close connection when control+C/cmd+C
	iris.RegisterOnInterrupt(func() {
		if err := c.redisConnection.Close(); err != nil {
			c.logger.Errorln(err)
		}
	})

	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: true,
		AllowReclaim:                true,
	})

	c.session.UseDatabase(c.redisConnection)
}
