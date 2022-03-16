package main

import (
	"fmt"
	"runtime"

	"github.com/containrrr/shoutrrr"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

func (c *Server) initLogging() {
	c.app.Logger().Handle(func(l *golog.Log) bool {
		_, fn, line, _ := runtime.Caller(5)

		message := l.Message
		source := fmt.Sprintf("%s#%d", fn, line)

		contextLogger := c.irisLogger.With(
			zap.String("file", source),
		)

		switch l.Level {
		case golog.DisableLevel:
			// do not log
		case golog.DebugLevel:
			contextLogger.Debug(message)
		case golog.InfoLevel:
			contextLogger.Info(message)
		case golog.WarnLevel:
			contextLogger.Warn(message)
		case golog.ErrorLevel:
			contextLogger.Error(message)
		case golog.FatalLevel:
			contextLogger.Fatal(message)
		default:
			contextLogger.Info(message)
		}

		return true
	})
}

func (c *Server) auditLog(ctx iris.Context, message string, depth int) {
	username := "*anonymous*"
	userId := ""
	user, _ := c.getUser(ctx)
	if user != nil {
		username = user.Username
		userId = user.Uuid
	}

	c.auditLogger.With(
		zap.String("requestMethod", ctx.Method()),
		zap.String("requestPath", ctx.Path()),
		zap.String("user", username),
		zap.String("userID", userId),
	).Info(message)
}

func (c *Server) notificationMessage(ctx iris.Context, message string) {
	username := "*anonymous*"
	user, _ := c.getUser(ctx)
	if user != nil {
		username = fmt.Sprintf("%s (%s)", user.Username, user.Uuid)
	}
	message = fmt.Sprintf(c.config.App.Notification.Message, username, message)

	// send notification
	for _, url := range c.config.App.Notification.Channels {
		if url != "" {
			if err := shoutrrr.Send(url, message); err != nil {
				c.logger.Errorf("unable to send shoutrrr notification: %v", err.Error())
			}
		}
	}
}
