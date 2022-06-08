package main

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	iris "github.com/kataras/iris/v12"

	"devops-console/models"
	"devops-console/models/formdata"
	"devops-console/models/response"
)

const (
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

	event := pagerduty.V2Event{
		RoutingKey: c.config.Support.Pagerduty.IntegrationKey,
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
			Message: fmt.Sprintf("%v", pagerdutyResponse.Message),
		}
		c.responseJson(ctx, resp)
	} else {
		c.respondError(ctx, err)
		return
	}
}
