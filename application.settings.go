package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"

	"devops-console/models"
	"devops-console/models/formdata"
	"devops-console/models/response"
)

type ApplicationSettings struct {
	*Server
	vaultClient *keyvault.BaseClient
}

func (c *ApplicationSettings) Get(ctx iris.Context, user *models.User) {
	var ret models.SettingsOverall

	var wgFetch sync.WaitGroup
	var wgProcess sync.WaitGroup

	secretChannel := make(chan models.SettingSecret)

	ret.Settings = c.config.Settings
	ret.User = map[string]string{}
	ret.Team = map[string]map[string]string{}

	for _, setting := range c.config.Settings.User {
		ret.User[setting.Name] = c.getKeyvaultSecret(c.userSecretName(user, setting.Name))
	}

	// fetch secret settings (backgrounded)
	for _, team := range user.Teams {
		ret.Team[team.Name] = map[string]string{}
		for _, setting := range c.config.Settings.Team {
			wgFetch.Add(1)
			go func(teamName, settingName string) {
				defer wgFetch.Done()
				secretChannel <- models.SettingSecret{
					TeamName:    teamName,
					SettingName: settingName,
					Secret:      c.getKeyvaultSecret(c.teamSecretName(teamName, settingName)),
				}
			}(team.Name, setting.Name)
		}
	}

	// collect backgrounded fetches
	wgProcess.Add(1)
	go func() {
		defer wgProcess.Done()

		for settingSecret := range secretChannel {
			ret.Team[settingSecret.TeamName][settingSecret.SettingName] = settingSecret.Secret
		}
	}()

	// wait until all fetches are finished
	wgFetch.Wait()
	close(secretChannel)
	wgProcess.Wait()

	c.responseJson(ctx, ret)
}

func (c *ApplicationSettings) ApiUpdateUser(ctx iris.Context, user *models.User) {
	var err error

	formData := formdata.GeneralSettings{}
	err = ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// validation
	validationMessages := []string{}
	for _, setting := range c.config.Settings.User {
		if val, ok := formData[setting.Name]; ok {
			if !setting.Validation.Validate(val) {
				validationMessages = append(validationMessages, fmt.Sprintf("validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
			}
		}
	}

	if len(validationMessages) >= 1 {
		c.respondError(ctx, errors.New(strings.Join(validationMessages, "\n")))
		return
	}

	// set values
	for _, setting := range c.config.Settings.User {
		secretTags := map[string]*string{}
		for name, value := range setting.Tags {
			secretTags[name] = to.StringPtr(value)
		}

		secretTags["user"] = &user.Username
		secretTags["setting"] = &setting.Name

		if val, ok := formData[setting.Name]; ok {
			err = c.setKeyvaultSecret(
				c.userSecretName(user, setting.Name),
				val,
				secretTags,
			)

			if err != nil {
				c.respondError(ctx, fmt.Errorf("failed updating keyvault"))
				return
			}
		}
	}

	c.auditLog(ctx, "updated personal settings", 1)
	PrometheusActions.With(prometheus.Labels{"scope": "settings", "type": "updatePersonal"}).Inc()

	resp := response.GeneralMessage{
		Message: "updated personal settings",
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationSettings) ApiUpdateTeam(ctx iris.Context, user *models.User) {
	var err error

	team := ctx.Params().GetString("team")
	if team == "" {
		c.respondError(ctx, errors.New("invalid team"))
		return
	}
	// membership check
	if !user.IsMemberOf(team) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", team))
		return
	}

	formData := formdata.GeneralSettings{}
	err = ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// validation
	validationMessages := []string{}
	for _, setting := range c.config.Settings.Team {
		if val, ok := formData[setting.Name]; ok {
			if !setting.Validation.Validate(val) {
				validationMessages = append(validationMessages, fmt.Sprintf("validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
			}
		}
	}

	if len(validationMessages) >= 1 {
		c.respondError(ctx, errors.New(strings.Join(validationMessages, "\n")))
		return
	}

	// set values
	for _, setting := range c.config.Settings.Team {
		secretTags := map[string]*string{}

		for name, value := range setting.Tags {
			secretTags[name] = to.StringPtr(value)
		}

		secretTags["team"] = &team
		secretTags["setting"] = &setting.Name

		if val, ok := formData[setting.Name]; ok {
			err = c.setKeyvaultSecret(
				c.teamSecretName(team, setting.Name),
				val,
				secretTags,
			)

			if err != nil {
				c.respondError(ctx, fmt.Errorf("failed setting keyvault"))
				return
			}
		}
	}

	c.auditLog(ctx, fmt.Sprintf("updated team \"%s\" settings", team), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "settings", "type": "updateTeam"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("updated team \"%s\" settings", team),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationSettings) userSecretName(user *models.User, name string) string {
	return fmt.Sprintf("user---%s---%s", user.Username, name)
}

func (c *ApplicationSettings) teamSecretName(team, name string) string {
	return fmt.Sprintf("team---%s---%s", team, name)
}

func (c *ApplicationSettings) getKeyvaultClient() *keyvault.BaseClient {
	var err error
	var keyvaultAuth autorest.Authorizer

	if c.vaultClient == nil {
		keyvaultAuth, err = auth.NewAuthorizerFromEnvironmentWithResource("https://vault.azure.net")
		if err != nil {
			panic(err)
		}

		client := keyvault.New()
		client.Authorizer = keyvaultAuth

		c.vaultClient = &client
	}

	return c.vaultClient
}

func (c *ApplicationSettings) setKeyvaultSecret(secretName, secretValue string, tags map[string]*string) error {
	ctx := context.Background()
	enabled := secretValue != ""

	secretName = strings.Replace(secretName, "_", "-", -1)
	secretParamSet := keyvault.SecretSetParameters{}
	secretParamSet.Value = &secretValue

	secretAttributs := keyvault.SecretAttributes{}
	secretAttributs.Enabled = &enabled
	secretParamSet.SecretAttributes = &secretAttributs

	secretParamSet.Tags = tags

	client := c.getKeyvaultClient()
	_, err := client.SetSecret(ctx, c.config.Settings.Vault.Url, secretName, secretParamSet)

	return err
}

func (c *ApplicationSettings) getKeyvaultSecret(secretName string) (secretValue string) {
	var err error
	var secretBundle keyvault.SecretBundle
	ctx := context.Background()

	client := c.getKeyvaultClient()
	secretBundle, err = client.GetSecret(ctx, c.config.Settings.Vault.Url, secretName, "")

	if err == nil {
		secretValue = *secretBundle.Value
	}

	return
}
