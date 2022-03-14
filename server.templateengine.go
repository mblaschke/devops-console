package main

import (
	"reflect"
	"runtime"

	azureSdkVersion "github.com/Azure/azure-sdk-for-go/version"
	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/view"
	"go.uber.org/zap"
)

func (c *Server) initTemplateEngine() {
	contextLogger := c.logger.With(zap.String("setup", "templateEngine"))

	contextLogger.Infof("init jet template engine")

	c.tmpl = iris.Jet("./templates", ".jet")

	c.tmpl.AddVar("appVersion", gitTag)
	c.tmpl.AddVar("appVersionCommit", gitCommit)
	c.tmpl.AddVar("runtimeVersion", runtime.Version())
	c.tmpl.AddVar("irisVersion", iris.Version)
	c.tmpl.AddVar("azureSdkVersion", azureSdkVersion.Number)
	c.tmpl.AddVar("appConfig", c.config.App)

	c.tmpl.AddFunc("MainFeatureIsEnabled", func(args view.JetArguments) reflect.Value {
		main := args.Get(0).String()
		return reflect.ValueOf(c.config.App.MainFeatureIsEnabled(main))
	})

	c.tmpl.AddFunc("FeatureIsEnabled", func(args view.JetArguments) reflect.Value {
		main := args.Get(0).String()
		branch := args.Get(1).String()
		return reflect.ValueOf(c.config.App.FeatureIsEnabled(main, branch))
	})

	c.app.RegisterView(c.tmpl)
}
