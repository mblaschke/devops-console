package main

import "github.com/kataras/iris"

func (c *Server) initRoutes() {
	c.logger.Infof(" - init app routes")

	applicationKubernetes := ApplicationKubernetes{Server: c}
	applicationAzure := ApplicationAzure{Server: c}
	applicationSettings := ApplicationSettings{Server: c}
	applicationGeneral := ApplicationGeneral{Server: c}
	applicationAuth := ApplicationAuth{Server: c}

	c.app.Get("/", c.index)
	c.app.Post("/login", applicationAuth.Login)
	c.app.Get("/oauth", applicationAuth.LoginViaOauth)
	c.app.Get("/logout", applicationAuth.Logout)

	c.app.Get("/api/general/stats", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppStats) })
	c.app.Get("/api/app/config", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationGeneral.handleApiAppConfig) })

	c.app.Get("/general/settings", func(ctx iris.Context) { c.react(ctx, "Settings") })
	c.app.Get("/api/general/settings", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.Get) })
	c.app.Post("/api/general/settings/user", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateUser) })
	c.app.Post("/api/general/settings/team/{team:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationSettings.ApiUpdateTeam) })

	c.app.Get("/general/about", func(ctx iris.Context) { c.template(ctx, "Settings", "about.jet") })

	c.app.Get("/kubernetes/cluster", func(ctx iris.Context) { c.react(ctx, "Kubernetes Cluster") })
	c.app.Get("/api/kubernetes/cluster", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiCluster) })

	c.app.Get("/kubernetes/namespaces", func(ctx iris.Context) { c.react(ctx, "Kubernetes Namespaces") })
	c.app.Get("/api/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceList) })
	c.app.Post("/api/kubernetes/namespace", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceCreate) })
	c.app.Delete("/api/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceDelete) })
	c.app.Put("/api/kubernetes/namespace/{namespace:string}", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceUpdate) })
	c.app.Post("/api/kubernetes/namespace/{namespace:string}/reset", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.ApiNamespaceReset) })

	c.app.Get("/kubernetes/kubeconfig", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationKubernetes.Kubeconfig) })

	c.app.Get("/azure/resourcegroup", func(ctx iris.Context) { c.react(ctx, "Azure ResourceGroup") })
	c.app.Post("/api/azure/resourcegroup", func(ctx iris.Context) { c.ensureLoggedIn(ctx, applicationAzure.ApiResourceGroupCreate) })

	c.logger.Infof(" - init static file handler")

	c.app.HandleDir("/static", "./static", iris.DirOptions{
		IndexName: "/index.html",

		Gzip:     false,
		ShowList: false,
	})
}
