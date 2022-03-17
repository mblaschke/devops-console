package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"

	"devops-console/models/response"
)

func (c *Server) respondError(ctx iris.Context, err error) {
	c.handleError(ctx, err, false)
}

func (c *Server) respondErrorWithPenalty(ctx iris.Context, err error) {
	if opts.ErrorPunishmentThreshold >= 0 {
		s := c.startSession(ctx)

		// ignore new sessions
		errorCounter, errorCounterErr := s.GetInt64("__errorCounter")
		if errorCounterErr != nil {
			errorCounter = 0
		}

		if errorCounter >= opts.ErrorPunishmentThreshold {
			// counter threshold reached, PUNISH
			c.auditLog(ctx, "error threshold reached, punishing user by killing session, original error was: "+err.Error(), 2)
			err = fmt.Errorf("sorry, too many errors occurred. Your session was terminated, please login again")

			c.handleError(ctx, err, true)
			return
		} else {
			// increase counter
			errorCounter++
			s.Set("__errorCounter", errorCounter)
		}
	}

	c.handleError(ctx, err, false)
}

func (c *Server) handleError(ctx iris.Context, err error, logout bool) {
	c.handleErrorWithStatus(ctx, http.StatusBadRequest, err, logout)
}

func (c *Server) handleErrorWithStatus(ctx iris.Context, statusCode int, err error, logout bool) {
	message := fmt.Sprintf("Error: %v", err)
	c.auditLog(ctx, message, 1)

	if logout {
		c.destroySession(ctx)
		ctx.StatusCode(iris.StatusUnauthorized)
	} else {
		ctx.StatusCode(statusCode)
	}

	// clear X-CSRF-token header, make sure it's empty
	ctx.Header("X-CSRF-Token", "")

	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		// XHR error
		c.responseJson(ctx, response.GeneralMessage{
			Message: message,
		})
	} else {
		// Page error
		ctx.ViewData("ERROR_MESSAGE", message)

		if logout {
			ctx.ViewData("title", "Login")
			if err := ctx.View("login.jet"); err != nil {
				c.logger.Error(err)
			}
		} else {
			ctx.ViewData("title", "Error")
			if err := ctx.View("error.jet"); err != nil {
				c.logger.Error(err)
			}
		}
	}

	ctx.StopExecution()
	panic(ctx)
}
