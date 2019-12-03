package main

import (
	"github.com/go-openapi/strfmt"
	iris "github.com/kataras/iris/v12"
	alertmanager "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
)

type ApplicationAlertmanager struct {
	*Server
}

func (c *ApplicationAlertmanager) getClient(ctx iris.Context, name string) *alertmanager.Alertmanager {
	// todo: check access

	client, err := c.config.Alertmanager.GetAlertmanagerInstance(name)
	if err != nil {
		c.respondError(ctx, err)
	}

	return client
}

func (c *ApplicationAlertmanager) ApiAlertsList(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	filter := []string{}

	params := alert.NewGetAlertsParams()
	params.SetFilter(filter)
	alerts, err := client.Alert.GetAlerts(params)

	if err != nil {
		c.respondError(ctx, err)
	}

	// TODO: check access

	ctx.JSON(alerts.Payload)
}

func (c *ApplicationAlertmanager) ApiSilencesList(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	filter := []string{
		"active=true",
	}

	params := silence.NewGetSilencesParams()
	params.SetFilter(filter)
	silences, err := client.Silence.GetSilences(params)

	// TODO: check access

	if err != nil {
		c.respondError(ctx, err)
	}

	ctx.JSON(silences.Payload)
}

func (c *ApplicationAlertmanager) ApiSilencesDelete(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	getParams := silence.NewGetSilenceParams()
	getParams.SilenceID = strfmt.UUID(ctx.Params().GetString("silence"))
	silenceResp, err := client.Silence.GetSilence(getParams)
	if err != nil {
		c.respondError(ctx, err)
	}

	// TODO: check access

	deleteParams := silence.NewDeleteSilenceParams()
	deleteParams.SilenceID = strfmt.UUID(*silenceResp.Payload.ID)
	if _, err := client.Silence.DeleteSilence(deleteParams); err != nil {
		c.respondError(ctx, err)
	}

	ctx.JSON("true")
}
