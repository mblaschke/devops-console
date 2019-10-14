package main

import (
	"context"
	"devops-console/models"
	"devops-console/models/formdata"
	"devops-console/models/response"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/kataras/iris"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
)

type ApplicationSettings struct {
	*Server
	vaultClient *keyvault.BaseClient
}

func (c *ApplicationSettings) Get(ctx iris.Context) {
	var ret models.SettingsOverall
	user := c.getUserOrStop(ctx)

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

	ctx.JSON(ret)
}

func (c *ApplicationSettings) ApiUpdateUser(ctx iris.Context) {
	var err error
	user := c.getUserOrStop(ctx)

	formData := formdata.GeneralSettings{}
	err = ctx.ReadJSON(&formData)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	// validation
	validationMessages := []string{}
	for _, setting := range c.config.Settings.User {
		if val, ok := formData[setting.Name]; ok {
			if !setting.Validation.Validate(val) {
				validationMessages = append(validationMessages, fmt.Sprintf("Validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
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
			secretTags[name] = &value
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
				c.respondError(ctx, errors.New(fmt.Sprintf("Failed setting keyvault")))
				return
			}
		}
	}

	c.auditLog(ctx, "Updated personal settings")
	PrometheusActions.With(prometheus.Labels{"scope": "settings", "type": "updatePersonal"}).Inc()

	resp := response.GeneralMessage{
		Message: "Updated personal settings",
	}


	ctx.JSON(resp)
}

func (c *ApplicationSettings) ApiUpdateTeam(ctx iris.Context) {
	var err error
	user := c.getUserOrStop(ctx)

	team := ctx.Params().GetString("team")
	if team == "" {
		c.respondError(ctx, errors.New("Invalid team"))
		return
	}
	// membership check
	if !user.IsMemberOf(team) {
		c.respondError(ctx, errors.New(fmt.Sprintf("Access to team \"%s\" denied", team)))
		return
	}

	formData := formdata.GeneralSettings{}
	err = ctx.ReadJSON(&formData)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	// validation
	validationMessages := []string{}
	for _, setting := range c.config.Settings.Team {
		if val, ok := formData[setting.Name]; ok {
			if !setting.Validation.Validate(val) {
				validationMessages = append(validationMessages, fmt.Sprintf("Validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
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
			secretTags[name] = &value
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
				c.respondError(ctx, errors.New(fmt.Sprintf("Failed setting keyvault")))
				return
			}
		}
	}

	c.auditLog(ctx, fmt.Sprintf("Updated team \"%s\" settings", team))
	PrometheusActions.With(prometheus.Labels{"scope": "settings", "type": "updateTeam"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("Updated team \"%s\" settings", team),
	}

	ctx.JSON(resp)
}

func (c *ApplicationSettings) userSecretName(user *models.User, name string) string {
	return fmt.Sprintf("user---%s---%s", user.Username, name)
}

func (c *ApplicationSettings) teamSecretName(team, name string) string {
	return fmt.Sprintf("team---%s---%s", team, name)
}

func (c *ApplicationSettings) getKeyvaultClient(vaultUrl string) *keyvault.BaseClient {
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

	client := c.getKeyvaultClient("")
	_, err := client.SetSecret(ctx, c.config.Settings.Vault.Url, secretName, secretParamSet)

	return err
}

func (c *ApplicationSettings) getKeyvaultSecret(secretName string) (secretValue string) {
	var err error
	var secretBundle keyvault.SecretBundle
	ctx := context.Background()

	client := c.getKeyvaultClient("")
	secretBundle, err = client.GetSecret(ctx, c.config.Settings.Vault.Url, secretName, "")

	if err == nil {
		secretValue = *secretBundle.Value
	}

	return
}
