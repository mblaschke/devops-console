package main

import (
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"runtime"
)

func (c *Server) initLogging() {
	c.app.Logger().Handle(func(l *golog.Log) bool {
		_, fn, line, _ := runtime.Caller(5)

		message := l.Message
		source := fmt.Sprintf("%s#%d", fn, line)

		contextLogger := log.With(
			zap.String("file", source),
		)

		switch l.Level {
		case golog.DisableLevel:
		case golog.DebugLevel:
			contextLogger.Debug(l.Message)
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
	user, _ := c.getUser(ctx)
	if user != nil {
		username = fmt.Sprintf("%s (%s)", user.Username, user.Uuid)
	}

	c.auditLogger.With(
		zap.String("context", "audit"),
		zap.String("user", username),
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
		if err := shoutrrr.Send(url, message); err != nil {
			c.logger.Errorf("unable to send shoutrrr notification: %v", err.Error())
		}
	}
}
