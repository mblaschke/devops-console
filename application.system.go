package main

import (
	"github.com/kataras/iris/v12"
)

type ApplicationSystem struct {
	*Server
}

func (c *Server) Healthz(ctx iris.Context) {
	if c.redisConnection != nil {
		redisSuccess, redisError := c.redisConnection.Config().Driver.PingPong()
		if redisError != nil {
			c.logger.Errorln("Healthz: Redis error: ", redisSuccess)
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Text("Error")
			return
		}

		if !redisSuccess {
			c.logger.Errorln("Healthz: Redis pingPong failed")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Text("Error")
			return
		}
	}

	ctx.Text("OK")
}
