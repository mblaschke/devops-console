package main

import (
	"errors"
	"time"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"go.uber.org/zap"
)

const (
	httpHeaderCsrfToken = "X-CSRF-Token" // #nosec
)

func (c *Server) initRoutes() {
	contextLogger := c.logger.With(zap.String("setup", "routes"))

	c.app.Use(c.before)
	requestLogger := logger.New(logger.Config{
		// Status displays status code
		Status: true,
		// IP displays request's remote address
		IP: true,
		// Method displays the http method
		Method: true,
		// Path displays the request path
		Path: true,
		// Query appends the url query to the Path.
		Query: opts.Debug,
		// if !empty then its contents derives from `ctx.Values().Get("logger_message")
		// will be added to the logs.
		MessageContextKeys: []string{"userIdentification"},

		LogFunc: func(endTime time.Time, latency time.Duration, status, ip, method, path string, message interface{}, headerMessage interface{}) {
			contextLogger := c.logger.With(
				zap.String("type", "request"),
				zap.Float64("latency", latency.Seconds()),
				zap.String("status", status),
				zap.String("ip", ip),
				zap.String("method", method),
				zap.String("path", path),
				zap.Any("context", message),
			)

			contextLogger.Info()
		},
	})

	contextLogger.Infof("init static file handler")

	staticParty := c.app.Party("/", c.defaultHeaders)
	staticParty.HandleDir("/static", "./static", iris.DirOptions{
		IndexName: "/index.html",

		Gzip:     false,
		ShowList: false,
	})
	c.app.Favicon("./static/img/favicon.ico")

	contextLogger.Infof("init app routes")

	applicationKubernetes := ApplicationKubernetes{Server: c}
	applicationAlertmanager := ApplicationAlertmanager{Server: c}
	applicationAzure := ApplicationAzure{Server: c}
	applicationSettings := ApplicationSettings{Server: c}
	applicationGeneral := ApplicationGeneral{Server: c}
	applicationConfig := ApplicationConfig{Server: c}
	applicationAuth := ApplicationAuth{Server: c}
	applicationSystem := ApplicationSystem{Server: c}
	applicationIndex := ApplicationIndex{Server: c}
	applicationSupport := ApplicationSupport{Server: c}

	publicParty := c.app.Party("/", requestLogger, c.defaultHeaders)
	{
		publicParty.Get("/", c.index)
		publicParty.Post("/login", applicationAuth.Login)
		publicParty.Get("/oauth", applicationAuth.LoginViaOauth)
		publicParty.Get("/logout", applicationAuth.Logout)
		publicParty.Get("/logout/forced", applicationAuth.LogoutForced)
	}

	healthParty := c.app.Party("/", c.defaultHeaders)
	{
		healthParty.Get("/_healthz", applicationSystem.Healthz)
	}

	pageParty := c.app.Party("/", requestLogger, c.defaultHeaders, c.csrfProtectionReferer, c.csrfProtectionRegenrateToken)
	{
		pageParty.Get("/home", func(ctx iris.Context) { c.home(ctx) })

		if c.config.App.FeatureIsEnabled("general", "settings") {
			pageParty.Get("/general/settings", func(ctx iris.Context) { c.react(ctx, "Settings") })
		}

		if c.config.App.FeatureIsEnabled("general", "about") {
			pageParty.Get("/general/about", func(ctx iris.Context) { c.template(ctx, "About", "about.jet") })
		}
		if c.config.App.FeatureIsEnabled("kubernetes", "namespaces") {
			pageParty.Get("/kubernetes/namespaces", func(ctx iris.Context) { c.react(ctx, "Kubernetes Namespaces") })
		}

		if c.config.App.FeatureIsEnabled("kubernetes", "access") {
			pageParty.Get("/kubernetes/access", func(ctx iris.Context) { c.react(ctx, "Kubernetes Kubeconfig") })
		}

		if c.config.App.FeatureIsEnabled("azure", "resourcegroups") {
			pageParty.Get("/azure/resourcegroup", func(ctx iris.Context) { c.react(ctx, "Azure ResourceGroup") })
		}

		if c.config.App.FeatureIsEnabled("azure", "roleassignments") {
			pageParty.Get("/azure/roleassignment", func(ctx iris.Context) { c.react(ctx, "Azure RoleAssignment") })
		}

		if c.config.App.FeatureIsEnabled("support", "pagerduty") {
			pageParty.Get("/support/pagerduty", func(ctx iris.Context) { c.react(ctx, "PagerDuty support") })
		}

		if c.config.App.FeatureIsEnabled("monitoring", "alertmanagers") {
			pageParty.Get("/monitoring/alertmanager", func(ctx iris.Context) { c.react(ctx, "Alertmanager") })
		}
	}

	apiParty := c.app.Party("/_webapi", requestLogger, c.defaultHeaders, c.csrfProtectionReferer, c.csrfProtectionToken)
	{
		apiParty.Get("/heartbeat", applicationIndex.heartbeat)

		if c.config.App.FeatureIsEnabled("general", "about") {
			apiParty.Get("/general/stats", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppStats) })
		}

		apiParty.Get("/app/config", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationConfig.handleApiAppConfig) })

		if c.config.App.FeatureIsEnabled("general", "settings") {
			apiParty.Get("/general/settings", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.Get) })
			apiParty.Post("/general/settings/user", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateUser) })
			apiParty.Post("/general/settings/team/{team:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateTeam) })
		}

		if c.config.App.FeatureIsEnabled("kubernetes", "namespaces") {
			apiParty.Get("/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceList) })
			apiParty.Post("/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceCreate) })
			apiParty.Delete("/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceDelete) })
			apiParty.Put("/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceUpdate) })
			apiParty.Post("/kubernetes/namespace/{namespace:string}/reset", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceReset) })
		}

		if c.config.App.FeatureIsEnabled("kubernetes", "access") {
			apiParty.Get("/kubernetes/kubeconfig", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.Kubeconfig) })
			apiParty.Get("/kubernetes/kubeconfig/{name:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.KubeconfigDownload) })
		}

		if c.config.App.FeatureIsEnabled("azure", "resourcegroups") {
			apiParty.Post("/azure/resourcegroup", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAzure.ApiResourceGroupCreate) })
		}

		if c.config.App.FeatureIsEnabled("azure", "roleassignments") {
			apiParty.Post("/azure/roleassignment", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAzure.ApiRoleAssignmentCreate) })
			apiParty.Delete("/azure/roleassignment", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAzure.ApiRoleAssignmentDelete) })
		}

		if c.config.App.FeatureIsEnabled("support", "pagerduty") {
			apiParty.Post("/support/pagerduty", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSupport.ApiPagerDutyTicketCreate) })
		}

		if c.config.App.FeatureIsEnabled("monitoring", "alertmanagers") {
			apiParty.Get("/alertmanager/{instance:string}/alerts", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiAlertsList) })
			apiParty.Get("/alertmanager/{instance:string}/silences", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesList) })
			apiParty.Delete("/alertmanager/{instance:string}/silence/{silence:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesDelete) })
			apiParty.Post("/alertmanager/{instance:string}/silence", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesCreate) })
			apiParty.Put("/alertmanager/{instance:string}/silence/{silence:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAlertmanager.ApiSilencesUpdate) })
		}
	}

	apiServiceParty := c.app.Party("/api", requestLogger, c.defaultHeaders)
	{
		if c.config.App.FeatureIsEnabled("kubernetes", "namespaces") {
			apiServiceParty.Post("/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureServiceUser(ctx, applicationKubernetes.ApiServiceNamespaceEnsure) })
		}
	}

	c.app.OnErrorCode(iris.StatusNotFound, c.defaultHeaders, c.notFound)
}

func (c *Server) notFound(ctx iris.Context) {
	defer func() {
		recover() //nolint:golint,errcheck
	}()

	c.respondErrorWithPenalty(ctx, errors.New("document not found"))
}

func (c *Server) before(ctx iris.Context) {
	// view information
	ctx.ViewData("navigationRoute", ctx.GetCurrentRoute().Path())
	ctx.Next()
}
