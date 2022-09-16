package main

import (
	"errors"
	"fmt"

	"github.com/go-openapi/strfmt"
	iris "github.com/kataras/iris/v12"
	alertmanager "github.com/prometheus/alertmanager/api/v2/client"
	alertmanagerAlert "github.com/prometheus/alertmanager/api/v2/client/alert"
	alertmanagerSilence "github.com/prometheus/alertmanager/api/v2/client/silence"
	alertmanagerModels "github.com/prometheus/alertmanager/api/v2/models"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/formdata"
	"github.com/mblaschke/devops-console/models/response"
)

type ApplicationAlertmanager struct {
	*Server
}

func (c *ApplicationAlertmanager) getClient(ctx iris.Context, name string) *alertmanager.Alertmanager {
	client, err := c.config.Alertmanager.GetAlertmanagerInstance(name)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
	}

	return client
}

func (c *ApplicationAlertmanager) ApiAlertsList(ctx iris.Context, user *models.User) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	filter := []string{}

	params := alertmanagerAlert.NewGetAlertsParams()
	params.SetFilter(filter)
	alerts, err := client.Alert.GetAlerts(params)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	alerts = c.filterAlerts(ctx, alerts)

	c.responseJson(ctx, alerts.Payload)
}

func (c *ApplicationAlertmanager) ApiSilencesList(ctx iris.Context, user *models.User) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	filter := []string{}

	params := alertmanagerSilence.NewGetSilencesParams()
	params.SetFilter(filter)
	silences, err := client.Silence.GetSilences(params)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	silences = c.filterSilences(ctx, silences)

	c.responseJson(ctx, silences.Payload)
}

func (c *ApplicationAlertmanager) ApiSilencesDelete(ctx iris.Context, user *models.User) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	getParams := alertmanagerSilence.NewGetSilenceParams()
	getParams.SilenceID = strfmt.UUID(ctx.Params().GetString("silence"))
	silenceResp, err := client.Silence.GetSilence(getParams)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if !c.checkSilenceAccess(ctx, user, silenceResp.Payload.Matchers) {
		c.respondErrorWithPenalty(ctx, errors.New("access to silence denied"))
		return
	}

	deleteParams := alertmanagerSilence.NewDeleteSilenceParams()
	deleteParams.SilenceID = strfmt.UUID(*silenceResp.Payload.ID)
	if _, err := client.Silence.DeleteSilence(deleteParams); err != nil {
		c.respondError(ctx, err)
		return
	}

	team := ""
	if val := c.getSilenceMatcherTeam(ctx, user, silenceResp.Payload.Matchers); val != nil {
		team = *val
	}
	c.notificationMessage(ctx, fmt.Sprintf("alertmanager silence %s for team \"%v\" deleted", *silenceResp.Payload.ID, team))

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("silence \"%s\" deleted", *silenceResp.Payload.ID),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationAlertmanager) ApiSilencesUpdate(ctx iris.Context, user *models.User) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	getParams := alertmanagerSilence.NewGetSilenceParams()
	getParams.SilenceID = strfmt.UUID(ctx.Params().GetString("silence"))
	silenceResp, err := client.Silence.GetSilence(getParams)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if !c.checkSilenceAccess(ctx, user, silenceResp.Payload.Matchers) {
		c.respondErrorWithPenalty(ctx, errors.New("access to silence denied"))
		return
	}

	formData := c.getSilenceFormData(ctx)
	formData.Silence.CreatedBy = silenceResp.Payload.CreatedBy

	postParams := alertmanagerSilence.NewPostSilencesParams()
	postParams.Silence = &alertmanagerModels.PostableSilence{
		ID:      *silenceResp.Payload.ID,
		Silence: formData.Silence,
	}
	if _, err := client.Silence.PostSilences(postParams); err != nil {
		c.respondError(ctx, err)
		return
	}

	c.notificationMessage(ctx, fmt.Sprintf("alertmanager silence %s for team \"%v\" updated", *silenceResp.Payload.ID, formData.Team))

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("silence \"%s\" updated", *silenceResp.Payload.ID),
	}

	c.responseJson(ctx, resp)

}

func (c *ApplicationAlertmanager) ApiSilencesCreate(ctx iris.Context, user *models.User) {
	client := c.getClient(ctx, ctx.Params().GetString("instance"))

	formData := c.getSilenceFormData(ctx)

	username := fmt.Sprintf("%v [%v]", user.Username, user.Uuid)
	formData.Silence.CreatedBy = &username

	postParams := alertmanagerSilence.NewPostSilencesParams()
	postParams.Silence = &alertmanagerModels.PostableSilence{
		Silence: formData.Silence,
	}

	if !c.checkSilenceAccess(ctx, user, postParams.Silence.Matchers) {
		c.respondErrorWithPenalty(ctx, errors.New("access to silence denied"))
		return
	}

	silenceResp, err := client.Silence.PostSilences(postParams)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	c.notificationMessage(ctx, fmt.Sprintf("alertmanager silence %s for team \"%v\" create", silenceResp.Payload.SilenceID, formData.Team))

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("silence \"%s\" created", silenceResp.Payload.SilenceID),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationAlertmanager) getSilenceFormData(ctx iris.Context) *formdata.AlertmanagerForm {
	var err error

	formData := &formdata.AlertmanagerForm{}
	if err := ctx.ReadJSON(&formData); err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return nil
	}

	// validate
	formData, err = formData.Validate()
	if err != nil {
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
		Name:    &label,
		Value:   &formData.Team,
		IsRegex: &labelIsRegexp,
	})

	formData.Silence.Matchers = matchers

	return formData
}

func (c *ApplicationAlertmanager) filterAlerts(ctx iris.Context, alerts *alertmanagerAlert.GetAlertsOK) *alertmanagerAlert.GetAlertsOK {
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

func (c *ApplicationAlertmanager) filterSilences(ctx iris.Context, silences *alertmanagerSilence.GetSilencesOK) *alertmanagerSilence.GetSilencesOK {
	user := c.getUserOrStop(ctx)

	filteredSilences := alertmanagerModels.GettableSilences{}

	for _, row := range silences.Payload {
		if c.checkSilenceAccess(ctx, user, row.Matchers) {
			filteredSilences = append(filteredSilences, row)
		}
	}
	silences.Payload = filteredSilences

	return silences
}

func (c *ApplicationAlertmanager) getSilenceMatcherTeam(ctx iris.Context, user *models.User, matchers alertmanagerModels.Matchers) (team *string) {
	for _, matcher := range matchers {
		if matcher.Name != nil && matcher.Value != nil {
			if *matcher.Name == "team" {
				team = matcher.Value
			}
		}
	}

	return
}

func (c *ApplicationAlertmanager) checkSilenceAccess(ctx iris.Context, user *models.User, matchers alertmanagerModels.Matchers) (status bool) {
	status = false

	team := c.getSilenceMatcherTeam(ctx, user, matchers)
	if team != nil && user.IsMemberOf(*team) {
		status = true
	}

	return
}
