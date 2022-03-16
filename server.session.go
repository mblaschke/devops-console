package main

import (
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"go.uber.org/zap"
)

const (
	SessionVarNameAppVersion = "__APPVERSION__"
	SessionVarNameUserAgent  = "__USERAGENT__"
)

func (c *Server) startSession(ctx iris.Context, cookieOptions ...context.CookieOption) *sessions.Session {
	cookieOptionList := []context.CookieOption{
		func(cookie *http.Cookie) {
			c.applySessionSetting(cookie)
		},
	}
	cookieOptionList = append(cookieOptionList, cookieOptions...)
	s := c.session.Start(ctx, cookieOptionList...)
	invalidSession := false

	// invalidate sessions for different app versions
	if !s.IsNew() {
		if val := s.GetString(SessionVarNameAppVersion); val != "" && val != gitTag {
			// session is invalid
			invalidSession = true
		}
	}

	// enforce same useragent
	userAgent := ctx.Request().UserAgent()
	if !s.IsNew() {
		if val := s.GetString(SessionVarNameUserAgent); val != "" && val != userAgent {
			// session is invalid
			invalidSession = true
		}
	}

	if invalidSession {
		c.session.Destroy(ctx)
		s = c.session.Start(ctx)
	}

	s.Set(SessionVarNameAppVersion, gitTag)
	s.Set(SessionVarNameUserAgent, userAgent)

	return s
}

func (c *Server) recreateSession(ctx iris.Context, cookieOptions ...context.CookieOption) *sessions.Session {
	c.session.Destroy(ctx)
	c.session.Destroy(ctx)

	return c.startSession(ctx, cookieOptions...)
}

func (c *Server) destroySession(ctx iris.Context) {
	c.session.Destroy(ctx)
}

func (c *Server) applySessionSetting(cookie *http.Cookie) {
	cookie.Secure = c.config.App.Session.CookieSecure

	switch strings.ToLower(c.config.App.Session.CookieSameSite) {
	case "default":
		cookie.SameSite = http.SameSiteDefaultMode
	case "lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "none":
		cookie.SameSite = http.SameSiteNoneMode
	}

	if c.config.App.Session.CookieDomain != "" {
		cookie.Domain = c.config.App.Session.CookieDomain
	}
}

func (c *Server) initSession() {
	contextLogger := c.logger.With(zap.String("setup", "session"))
	contextLogger.Infof("using %v session", c.config.App.Session.Type)
	contextLogger.Infof("cookie name: %v", c.config.App.Session.CookieName)
	contextLogger.Infof("session expiry: %v", c.config.App.Session.Expiry)

	switch c.config.App.Session.Type {
	case "internal":
		c.initSessionInternal()
	case "securecookie":
		c.initSessionSecureCookie()
	case "redis":
		c.initSessionRedis()
	case "redis+securecookie":
		c.initSessionRedisSecureCookie()
	default:
		panic("invalid session type defined")
	}
}

func (c *Server) initSessionInternal() {
	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		CookieSecureTLS:             c.config.App.Session.CookieSecure,
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
		CookieSecureTLS:             c.config.App.Session.CookieSecure,
		Encode:                      secureCookie.Encode,
		Decode:                      secureCookie.Decode,
		AllowReclaim:                true,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: true,
	})
}

func (c *Server) createRedisConnection() {
	contextLogger := c.logger.With(
		zap.String("setup", "session"),
		zap.String("session", "redis"),
	)

	// try connect to redis server (with retry)
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

		contextLogger.Errorf("redis connection failed, retrying in %v", retryTime.String())
		time.Sleep(retryTime)
	}

	if c.redisConnection == nil {
		contextLogger.Fatal("redis connection failed, cannot connect to session database")
	}

	contextLogger.Info("redis connection established")

	// Close connection when control+C/cmd+C
	iris.RegisterOnInterrupt(func() {
		if err := c.redisConnection.Close(); err != nil {
			contextLogger.Error(err)
		}
	})
}

func (c *Server) initSessionRedis() {
	c.createRedisConnection()

	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		CookieSecureTLS:             c.config.App.Session.CookieSecure,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: true,
		AllowReclaim:                true,
	})
	c.session.UseDatabase(c.redisConnection)
}

func (c *Server) initSessionRedisSecureCookie() {
	c.createRedisConnection()

	secureCookie := securecookie.New(
		[]byte(c.config.App.Session.SecureCookie.HashKey),
		[]byte(c.config.App.Session.SecureCookie.BlockKey),
	)

	c.session = sessions.New(sessions.Config{
		Cookie:                      c.config.App.Session.CookieName,
		CookieSecureTLS:             c.config.App.Session.CookieSecure,
		Encode:                      secureCookie.Encode,
		Decode:                      secureCookie.Decode,
		AllowReclaim:                true,
		Expires:                     c.config.App.Session.Expiry,
		DisableSubdomainPersistence: true,
	})
	c.session.UseDatabase(c.redisConnection)
}
