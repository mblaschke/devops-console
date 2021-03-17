package main

import (
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/kataras/iris/v12"
)

func (c *Server) auditLog(ctx iris.Context, message string, depth int) {
	username := "*anonymous*"
	user, _ := c.getUser(ctx)
	if user != nil {
		username = fmt.Sprintf("%s (%s)", user.Username, user.Uuid)
	}

	c.logger.InfoDepth(depth, fmt.Sprintf("AUDIT: user[%s]: %s", username, message))
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
			c.logger.Errorf("Unable to send shoutrrr notification: %v", err.Error())
		}
	}
}
