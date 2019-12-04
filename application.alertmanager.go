package main

import (
	"devops-console/models"
	"devops-console/models/formdata"
	"errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	iris "github.com/kataras/iris/v12"
	alertmanager "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	alertmanagerModels "github.com/prometheus/alertmanager/api/v2/models"
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

	alerts = c.filterAlerts(ctx, alerts)

	ctx.JSON(alerts.Payload)
}

func (c *ApplicationAlertmanager) ApiSilencesList(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	filter := []string{}

	params := silence.NewGetSilencesParams()
	params.SetFilter(filter)
	silences, err := client.Silence.GetSilences(params)

	if err != nil {
		c.respondError(ctx, err)
	}

	silences = c.filterSilences(ctx, silences)

	ctx.JSON(silences.Payload)
}

func (c *ApplicationAlertmanager) ApiSilencesDelete(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))
	user := c.getUserOrStop(ctx)

	getParams := silence.NewGetSilenceParams()
	getParams.SilenceID = strfmt.UUID(ctx.Params().GetString("silence"))
	silenceResp, err := client.Silence.GetSilence(getParams)
	if err != nil {
		c.respondError(ctx, err)
	}

	if !c.checkSilenceAccess(ctx, user,  silenceResp.Payload) {
		c.respondError(ctx, errors.New("Access to silence denied"))
	}

	deleteParams := silence.NewDeleteSilenceParams()
	deleteParams.SilenceID = strfmt.UUID(*silenceResp.Payload.ID)
	if _, err := client.Silence.DeleteSilence(deleteParams); err != nil {
		c.respondError(ctx, err)
	}

	ctx.JSON("true")
}

func (c *ApplicationAlertmanager) ApiSilencesUpdate(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))
	user := c.getUserOrStop(ctx)

	getParams := silence.NewGetSilenceParams()
	getParams.SilenceID = strfmt.UUID(ctx.Params().GetString("silence"))
	silenceResp, err := client.Silence.GetSilence(getParams)
	if err != nil {
		c.respondError(ctx, err)
	}

	if !c.checkSilenceAccess(ctx, user,  silenceResp.Payload) {
		c.respondError(ctx, errors.New("Access to silence denied"))
	}

	formData := c.getSilenceFormData(ctx)

	postParams := silence.NewPostSilencesParams()
	postParams.Silence = &alertmanagerModels.PostableSilence{
		ID: *silenceResp.Payload.ID,
		Silence: formData.Silence,
	}
	if _, err := client.Silence.PostSilences(postParams); err != nil {
		c.respondError(ctx, err)
	}

	ctx.JSON("true")
}

func (c *ApplicationAlertmanager) ApiSilencesCreate(ctx iris.Context) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))
	user := c.getUserOrStop(ctx)

	formData := c.getSilenceFormData(ctx)

	username := fmt.Sprintf("%v [%v]", user.Username, user.Uuid)
	formData.Silence.CreatedBy = &username

	postParams := silence.NewPostSilencesParams()
	postParams.Silence = &alertmanagerModels.PostableSilence{
		Silence: formData.Silence,
	}
	if _, err := client.Silence.PostSilences(postParams); err != nil {
		c.respondError(ctx, err)
	}

	ctx.JSON("true")
}

func (c *ApplicationAlertmanager) getSilenceFormData(ctx iris.Context) *formdata.AlertmanagerForm {
	formData := &formdata.AlertmanagerForm{}
	if err := ctx.ReadJSON(&formData); err != nil {
		c.respondError(ctx, err)
		return nil
	}

	// filter team
	matchers := alertmanagerModels.Matchers{}
	for _, row := range formData.Silence.Matchers {
		if *row.Name != "team" {
			matchers = append(matchers, row)
		}
	}

	// add team
	label := "team"
	labelIsRegexp := false
	matchers = append(matchers, &alertmanagerModels.Matcher{
		Name: &label,
		Value: &formData.Team,
		IsRegex: &labelIsRegexp,
	})

	formData.Silence.Matchers = matchers

	return formData
}

func (c *ApplicationAlertmanager) filterAlerts(ctx iris.Context, alerts *alert.GetAlertsOK) *alert.GetAlertsOK {
	user := c.getUserOrStop(ctx)

	filteredAlerts := alertmanagerModels.GettableAlerts{}

	for _, row := range alerts.Payload {
		if c.checkAlertAccess(ctx, user, row) {
			filteredAlerts = append(filteredAlerts, row)
		}
	}
	alerts.Payload = filteredAlerts

	return alerts
}

func (c *ApplicationAlertmanager) checkAlertAccess(ctx iris.Context, user *models.User, row *alertmanagerModels.GettableAlert) (status bool) {
	status = false

	if team, ok := row.Labels["team"]; ok {
		if user.IsMemberOf(team) {
			status = true
		}
	}

	return
}

func (c *ApplicationAlertmanager) filterSilences(ctx iris.Context, silences *silence.GetSilencesOK) *silence.GetSilencesOK {
	user := c.getUserOrStop(ctx)

	filteredSilences := alertmanagerModels.GettableSilences{}

	for _, row := range silences.Payload {
		if c.checkSilenceAccess(ctx, user, row) {
			filteredSilences = append(filteredSilences, row)
		}
	}
	silences.Payload = filteredSilences

	return silences
}

func (c *ApplicationAlertmanager) checkSilenceAccess(ctx iris.Context, user *models.User, row *alertmanagerModels.GettableSilence) (status bool) {
	status = false

	for _, matcher := range row.Matchers {
		if matcher.Name != nil && matcher.Value != nil {
			if *matcher.Name == "team" && user.IsMemberOf(*matcher.Value) {
				status = true
			}
		}
	}

	return
}
