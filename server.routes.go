package main

import (
	"encoding/json"
	"errors"
	iris "github.com/kataras/iris/v12"
)

func (c *Server) initRoutes() {
	c.app.Use(c.before)

	c.logger.Infof(" - init static file handler")

	staticParty := c.app.Party("/", c.defaultHeaders)
	staticParty.HandleDir("/static", "./static", iris.DirOptions{
		IndexName: "/index.html",

		Gzip:     false,
		ShowList: false,
	})

	c.logger.Infof(" - init app routes")

	applicationKubernetes := ApplicationKubernetes{Server: c}
	applicationAlertmanager := ApplicationAlertmanager{Server: c}
	applicationAzure := ApplicationAzure{Server: c}
	applicationSettings := ApplicationSettings{Server: c}
	applicationGeneral := ApplicationGeneral{Server: c}
	applicationAuth := ApplicationAuth{Server: c}

	publicParty := c.app.Party("/", c.defaultHeaders)
	{
		publicParty.Get("/", c.index)
		publicParty.Post("/login", applicationAuth.Login)
		publicParty.Get("/oauth", applicationAuth.LoginViaOauth)
		publicParty.Get("/logout", applicationAuth.Logout)
		publicParty.Get("/logout/forced", applicationAuth.LogoutForced)
	}

	pageParty := c.app.Party("/", c.defaultHeaders, c.csrfProtectionReferer, c.csrfProtectionToken, c.csrfProtectionRegenrateToken)
	{
		pageParty.Get("/general/settings", func(ctx iris.Context) { c.react(ctx, "Settings") })
		pageParty.Get("/general/about", func(ctx iris.Context) { c.template(ctx, "About", "about.jet") })
		pageParty.Get("/kubernetes/namespaces", func(ctx iris.Context) { c.react(ctx, "Kubernetes Namespaces") })
		pageParty.Get("/kubernetes/cluster", func(ctx iris.Context) { c.react(ctx, "Kubernetes Cluster") })
		pageParty.Get("/kubernetes/kubeconfig", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.Kubeconfig) })
		pageParty.Get("/azure/resourcegroup", func(ctx iris.Context) { c.react(ctx, "Azure ResourceGroup") })
		pageParty.Get("/monitoring/alertmanager", func(ctx iris.Context) { c.react(ctx, "Alertmanager") })
	}

	apiParty := c.app.Party("/api", c.defaultHeaders, c.csrfProtectionReferer, c.csrfProtectionToken)
	{
		apiParty.Get("/general/stats", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppStats) })
		apiParty.Get("/app/config", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppConfig) })

		apiParty.Get("/general/settings", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.Get) })
		apiParty.Post("/general/settings/user", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateUser) })
		apiParty.Post("/general/settings/team/{team:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateTeam) })
		
		apiParty.Get("/kubernetes/cluster", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiCluster) })

		apiParty.Get("/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceList) })
		apiParty.Post("/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceCreate) })
		apiParty.Delete("/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceDelete) })
		apiParty.Put("/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceUpdate) })
		apiParty.Post("/kubernetes/namespace/{namespace:string}/reset", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceReset) })

		apiParty.Post("/azure/resourcegroup", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAzure.ApiResourceGroupCreate) })
		
		apiParty.Get("/alertmanager/{instance:string}/alerts", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiAlertsList) })
		apiParty.Get("/alertmanager/{instance:string}/silences", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesList) })
		apiParty.Delete("/alertmanager/{instance:string}/silence/{silence:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesDelete) })
		apiParty.Post("/alertmanager/{instance:string}/silence", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesCreate) })
		apiParty.Put("/alertmanager/{instance:string}/silence/{silence:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesUpdate) })
	}

	c.app.OnErrorCode(iris.StatusNotFound, c.notFound)
}


func (c *Server) notFound(ctx iris.Context) {
	defer func() {
		recover()
	}()

	c.respondErrorWithPenalty(ctx, errors.New("Document not found"))
}

func (c *Server) before(ctx iris.Context) {
	// view information
	ctx.ViewData("navigationRoute", ctx.GetCurrentRoute().Path())
	ctx.Next()
}

func (c *Server) defaultHeaders(ctx iris.Context) {
	// security headers
	ctx.Header("X-Frame-Options", "BLOCK")
	ctx.Header("X-XSS-Protection", "1; mode=block")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")
	ctx.Next()
}

func (c *Server) csrfProtectionReferer(ctx iris.Context) {
	// TODO
	ctx.Next()
}

func (c *Server) csrfProtectionRegenrateToken(ctx iris.Context) {
	c.csrfProtectionTokenRegenerate(ctx)
	ctx.Next()
}

func (c *Server) csrfProtectionToken(ctx iris.Context) {
	if opts.DisableCsrfProtection {
		ctx.ViewData("CSRF_TOKEN_JSON", "false")
		ctx.Next()
		return
	}

	s := c.session.Start(ctx)

	// get token
	sessionToken := ""
	if val, ok := s.Get("CSRF").(string); ok {
		sessionToken = val
	}

	if sessionToken == "" {
		sessionToken = c.csrfProtectionTokenRegenerate(ctx)
	}

	method := ctx.Method()

	// check token if not GET or HEAD (safe methods)
	if method != "GET" && method != "HEAD" {
		clientToken := ctx.GetHeader("X-CSRF-Token")

		if sessionToken == "" || clientToken != sessionToken {
			c.respondErrorWithPenalty(ctx, errors.New("Invalid CSRF token"))
			return
		}
	}

	// inject token
	ctx.Header("X-CSRF-Token", sessionToken)

	tokenJson, _ := json.Marshal(sessionToken)
	ctx.ViewData("CSRF_TOKEN_JSON", tokenJson)

	ctx.Next()
}

func (c *Server) csrfProtectionTokenRegenerate(ctx iris.Context) string {
	if opts.DisableCsrfProtection {
		return ""
	}

	s := c.session.Start(ctx)

	// set new token
	token := randomString(64);
	s.Set("CSRF", token);
	ctx.Header("X-CSRF-Token", token)

	tokenJson, _ := json.Marshal(token)
	ctx.ViewData("CSRF_TOKEN_JSON", tokenJson)

	return token
}
