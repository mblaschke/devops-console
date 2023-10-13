package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	iris "github.com/kataras/iris/v12"
	redis "github.com/redis/go-redis/v9"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/formdata"
	"github.com/mblaschke/devops-console/models/response"
	"github.com/mblaschke/devops-console/services"
)

const (
	RedisPagerDutyEndpointList = `support:pagerduty:endpoint:list`
	RedisPagerDutyEndpointLock = `support:pagerduty:endpoint:lock`

	SupportPagerdutyEventComponent = `
Type: %v
Location: %v
Group/Namespace: %v
Resource: %v
`

	SupportPagerdutyEventSummary = `
Requested by:
%v for %v

Message:
%v

Contact:
%v
`
)

type ApplicationSupport struct {
	*Server
}

func NewApplicationSupport(c *Server) *ApplicationSupport {
	app := ApplicationSupport{Server: c}
	app.init()

	return &app
}

func (c *ApplicationSupport) init() {
	services.PagerDutySetStaticEndpoints(c.config)

	if c.config.Support.Pagerduty.AuthToken != "" {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Errorf(`failed to update pagerduty endpoints: %v`, r)
				}
			}()

			for {
				c.updatePagerDutyEndpoints()
				time.Sleep(1 * time.Minute)
			}
		}()
	}
}

func (c *ApplicationSupport) updatePagerDutyEndpoints() {
	var err error
	endpointList := map[string]models.PagerDutyEndpoint{}

	ctx := context.Background()

	forceUpdate := false

	err = c.redis.Get(ctx, RedisPagerDutyEndpointLock).Err()
	if errors.Is(err, redis.Nil) {
		forceUpdate = true
	}

	val, err := c.redis.Get(ctx, RedisPagerDutyEndpointList).Result()
	if err == nil {
		err := json.Unmarshal([]byte(val), &endpointList)
		if err == nil {
			c.logger.Infof(`updating PagerDuty endpoint list from redis, got %v endpoints`, len(endpointList))
			services.PagerDutySetEndpointList(endpointList)
		} else {
			forceUpdate = true
		}
	}

	if forceUpdate {
		c.logger.Info(`updating PagerDuty endpoint list`)
		services.PagerDutyUpdateEndpointList(ctx, c.config)
		endpointList := services.PagerDutyGetEndpointList()
		c.logger.Infof(`found %v PagerDuty endpoints`, len(endpointList))
		c.redis.Set(ctx, RedisPagerDutyEndpointLock, 1, 30*time.Minute)

		if val, err := json.Marshal(endpointList); err == nil {
			c.redis.Set(ctx, RedisPagerDutyEndpointList, val, 3000*time.Hour)
		}
	}
}

func (c *ApplicationSupport) ApiPagerDutyTicketCreate(ctx iris.Context, user *models.User) {
	var err error

	formData := &formdata.SupportPagerduty{}
	err = ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// validate
	formData, err = formData.Validate()
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	endpointList := services.PagerDutyGetEndpointList()
	if _, exists := endpointList[formData.Endpoint]; !exists {
		c.respondError(ctx, fmt.Errorf(`invalid endpoint selected`))
		return
	}

	event := pagerduty.V2Event{
		RoutingKey: endpointList[formData.Endpoint].RoutingKey,
		ClientURL:  c.config.Support.Pagerduty.ClientURL,
		Action:     "trigger",
		Payload: &pagerduty.V2Payload{
			Summary:  fmt.Sprintf("emergency support request from %v (request by %v)", formData.Team, user.Email),
			Source:   "DevOps console",
			Severity: "critical",
			Details: pagerduty.V2Payload{
				Group: formData.ResourceType,
				Component: fmt.Sprintf(
					SupportPagerdutyEventComponent,
					formData.ResourceType,
					formData.Location,
					formData.ResourceGroup,
					formData.Resource,
				),
				Summary: fmt.Sprintf(
					SupportPagerdutyEventSummary,
					fmt.Sprintf("%s (%s)", user.Username, user.Email),
					formData.Team,
					formData.Message,
					formData.Contact,
				),
			},
		},
	}

	if pagerdutyResponse, err := pagerduty.ManageEventWithContext(context.Background(), event); err == nil {
		resp := response.GeneralMessage{
			Message: fmt.Sprintf(`PagerDuty response: %v`, pagerdutyResponse.Message),
		}
		c.responseJson(ctx, resp)
	} else {
		c.respondError(ctx, err)
		return
	}
}
