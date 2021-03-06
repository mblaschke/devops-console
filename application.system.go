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
			c.logger.Error("healthz: redis error: ", redisSuccess)
			ctx.StatusCode(iris.StatusInternalServerError)
			if _, err := ctx.Text("Error"); err != nil {
				c.logger.Errorf("error while sending response: %v", err)
			}

			return
		}

		if !redisSuccess {
			c.logger.Error("healthz: redis pingPong failed")
			ctx.StatusCode(iris.StatusInternalServerError)
			if _, err := ctx.Text("Error"); err != nil {
				c.logger.Errorf("error while sending response: %v", err)
			}
			return
		}
	}

	if _, err := ctx.Text("OK"); err != nil {
		c.logger.Errorf("error while sending response: %v", err)
	}
}
