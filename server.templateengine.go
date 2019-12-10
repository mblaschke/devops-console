package main

import (
	azureSdkVersion "github.com/Azure/azure-sdk-for-go/version"
	iris "github.com/kataras/iris"
	"runtime"
)

func (c *Server) initTemplateEngine() {
	c.logger.Infof(" - init jet template engine")

	c.tmpl = iris.Jet("./templates", ".jet")

	c.tmpl.AddVar("appVersion", Version)
	c.tmpl.AddVar("runtimeVersion", runtime.Version())
	c.tmpl.AddVar("irisVersion", iris.Version)
	c.tmpl.AddVar("azureSdkVersion", azureSdkVersion.Number)

	c.app.RegisterView(c.tmpl)
}
