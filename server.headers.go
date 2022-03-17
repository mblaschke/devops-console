package main

import (
	"encoding/base64"
	"errors"
	"fmt"

	uuid "github.com/iris-contrib/go.uuid"
	"github.com/kataras/iris/v12"
)

func (c *Server) defaultHeaders(ctx iris.Context) {
	// fallback if nonce generation failed -> 'self'
	cspNonceValue := ""
	cspNoneHeader := "'self'"
	if nonce, err := uuid.NewV4(); err == nil {
		cspNonceValue = base64.StdEncoding.EncodeToString(nonce.Bytes())
		cspNoneHeader = fmt.Sprintf(`'nonce-%[1]s'`, cspNonceValue)
	}

	ctx.ViewData("CSP_NONCE", cspNonceValue)
	ctx.Header(
		"Content-Security-Policy",
		fmt.Sprintf(
			`default-src 'none'; connect-src 'self'; img-src 'self'; font-src 'self'; script-src 'none'; script-src-elem %[1]s; style-src-elem %[1]s;`,
			cspNoneHeader,
		),
	)

	// security headers
	ctx.Header("X-Frame-Options", "DENY")
	ctx.Header("X-XSS-Protection", "1; mode=block")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Next()
}

func (c *Server) csrfProtectionReferer(ctx iris.Context) {
	ctx.Next()
}

func (c *Server) csrfProtectionRegenrateToken(ctx iris.Context) {
	c.csrfProtectionTokenRegenerate(ctx)
	ctx.Next()
}

func (c *Server) csrfProtectionToken(ctx iris.Context) {
	if opts.DisableCsrfProtection {
		ctx.ViewData("CSRF_TOKEN_JSON", "")
		ctx.Next()
		return
	}

	s := c.startSession(ctx)

	// get token
	sessionToken := ""
	if val, ok := s.Get("CSRF").(string); ok {
		sessionToken = val
	}

	if sessionToken == "" {
		sessionToken = c.csrfProtectionTokenRegenerate(ctx)
	}

	method := ctx.Method()

	// check token if not HEAD (safe methods)
	if method != "HEAD" {
		clientToken := ctx.GetHeader(httpHeaderCsrfToken)

		if sessionToken == "" || clientToken != sessionToken {
			c.respondErrorWithPenalty(ctx, errors.New("invalid CSRF token"))
			return
		}
	}

	// ctx.Header(httpHeaderCsrfToken, sessionToken)

	ctx.ViewData("CSRF_TOKEN", sessionToken)

	ctx.Next()
}

func (c *Server) csrfProtectionTokenRegenerate(ctx iris.Context) string {
	if opts.DisableCsrfProtection {
		return ""
	}

	s := c.startSession(ctx)

	// set new token
	token := randomString(64)
	s.Set("CSRF", token)
	ctx.Header(httpHeaderCsrfToken, token)

	ctx.ViewData("CSRF_TOKEN", token)

	return token
}
