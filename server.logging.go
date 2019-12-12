package main

import (
	"bytes"
	"devops-console/models/notification"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"net/http"
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
	c.notificationMessageWithContext(ctx, message, nil)
}

func (c *Server) notificationMessageWithContext(ctx iris.Context, message string, context *string) {
	if c.config.App.Notification.Slack.Webhook == "" {
		return
	}

	username := "*anonymous*"
	user, _ := c.getUser(ctx)
	if user != nil {
		username = fmt.Sprintf("%s (%s)", user.Username, user.Uuid)
	}

	message = fmt.Sprintf(c.config.App.Notification.Slack.Message, username, message)
	if context != nil {
		message += fmt.Sprintf("\n\n%s", *context)
	}

	payloadBlocks := []notification.NotificationMessageBlockContext{}

	payloadBlocks = append(payloadBlocks, notification.NotificationMessageBlockContext{
		Type: "section",
		Text: &notification.NotificationMessageBlockText{
			Type: "plain_text",
			Text: message,
		},
	})

	payload := notification.NotificationMessage{
		Channel:  c.config.App.Notification.Slack.Channel,
		Username: "devops-console",
		Text:     message,
		Blocks:   payloadBlocks,
	}
	payloadJson, _ := json.Marshal(payload)

	client := http.Client{}
	req, err := http.NewRequest("POST", c.config.App.Notification.Slack.Webhook, bytes.NewBuffer(payloadJson))
	defer req.Body.Close()
	if err != nil {
		c.logger.Errorf("Failed to send slack notification: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to send slack notification: %v", err)
	}
}
