package main

import (
	"encoding/json"
	"errors"
	iris "github.com/kataras/iris/v12"
)

func (c *Server) initRoutes() {
	c.logger.Infof(" - init static file handler")

	c.app.HandleDir("/static", "./static", iris.DirOptions{
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


	c.app.Get("/", c.index)
	c.app.Post("/login", applicationAuth.Login)
	c.app.Get("/oauth", applicationAuth.LoginViaOauth)
	c.app.Get("/logout", applicationAuth.Logout)

	party := c.app.Party("/", c.csrfProtectionReferer, c.csrfProtectionToken)
	{
		party.Get("/api/general/stats", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppStats) })
		party.Get("/api/app/config", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppConfig) })

		party.Get("/general/settings", func(ctx iris.Context) { c.react(ctx, "Settings") })
		party.Get("/api/general/settings", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.Get) })
		party.Post("/api/general/settings/user", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateUser) })
		party.Post("/api/general/settings/team/{team:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateTeam) })

		party.Get("/general/about", func(ctx iris.Context) { c.template(ctx, "About", "about.jet") })

		party.Get("/kubernetes/cluster", func(ctx iris.Context) { c.react(ctx, "Kubernetes Cluster") })
		party.Get("/api/kubernetes/cluster", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiCluster) })

		party.Get("/kubernetes/namespaces", func(ctx iris.Context) { c.react(ctx, "Kubernetes Namespaces") })
		party.Get("/api/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceList) })
		party.Post("/api/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceCreate) })
		party.Delete("/api/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceDelete) })
		party.Put("/api/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceUpdate) })
		party.Post("/api/kubernetes/namespace/{namespace:string}/reset", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceReset) })

		party.Get("/kubernetes/kubeconfig", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.Kubeconfig) })

		party.Get("/azure/resourcegroup", func(ctx iris.Context) { c.react(ctx, "Azure ResourceGroup") })
		party.Post("/api/azure/resourcegroup", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAzure.ApiResourceGroupCreate) })

		party.Get("/monitoring/alertmanager", func(ctx iris.Context) { c.react(ctx, "Alertmanager") })

		party.Get("/api/alertmanager/{instance:string}/alerts", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiAlertsList) })
		party.Get("/api/alertmanager/{instance:string}/silences", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesList) })
		party.Delete("/api/alertmanager/{instance:string}/silence/{silence:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesDelete) })
		party.Post("/api/alertmanager/{instance:string}/silence", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesCreate) })
		party.Put("/api/alertmanager/{instance:string}/silence/{silence:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesUpdate) })
	}
}

func (c *Server) csrfProtectionReferer(ctx iris.Context) {
	// TODO
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
			c.respondError(ctx, errors.New("Invalid CSRF token"))
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
