package main

import (
	azureSdkVersion "github.com/Azure/azure-sdk-for-go/version"
	iris "github.com/kataras/iris/v12"
	"runtime"
)

func (c *Server) initTemplateEngine() {
	c.logger.Infof(" - init jet template engine")

	c.tmpl = iris.Jet("./templates", ".jet")

	c.tmpl.AddVar("appVersion", gitTag)
	c.tmpl.AddVar("appVersionCommit", gitCommit)
	c.tmpl.AddVar("runtimeVersion", runtime.Version())
	c.tmpl.AddVar("irisVersion", iris.Version)
	c.tmpl.AddVar("azureSdkVersion", azureSdkVersion.Number)

	c.app.RegisterView(c.tmpl)
}
