package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/formdata"
	"github.com/mblaschke/devops-console/models/response"
)

type ApplicationSettings struct {
	*Server

	armClient *armclient.ArmClient
}

func NewApplicationSettings(c *Server) *ApplicationSettings {
	app := ApplicationSettings{Server: c}

	armClient, err := armclient.NewArmClientWithCloudName(opts.Azure.Environment, logrus.StandardLogger())
	if err != nil {
		log.Panic(err.Error())
	}

	armClient.SetUserAgent(UserAgent + gitTag)

	app.armClient = armClient
	return &app
}

func (c *ApplicationSettings) Get(ctx iris.Context, user *models.User) {
	var (
		ret       models.SettingsOverall
		wgFetch   sync.WaitGroup
		wgProcess sync.WaitGroup
	)

	secretChannel := make(chan models.SettingSecret)

	secretClient := c.keyVaultClient(ctx, c.config.Settings.Vault.Url)

	ret.Settings = c.config.Settings
	ret.User = map[string]string{}
	ret.Team = map[string]map[string]string{}

	for _, setting := range c.config.Settings.User {
		ret.User[setting.Name] = c.getKeyvaultSecret(secretClient, c.userSecretName(user, setting.Name))
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
					Secret:      c.getKeyvaultSecret(secretClient, c.teamSecretName(teamName, settingName)),
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

	secretClient := c.keyVaultClient(ctx, c.config.Settings.Vault.Url)

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
				secretClient,
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

	secretClient := c.keyVaultClient(ctx, c.config.Settings.Vault.Url)

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
				secretClient,
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

func (c *ApplicationSettings) keyVaultClient(ctx iris.Context, vaultUrl string) *azsecrets.Client {
	secretOpts := azsecrets.ClientOptions{
		ClientOptions: *c.armClient.NewAzCoreClientOptions(),
	}

	client, err := azsecrets.NewClient(vaultUrl, c.armClient.GetCred(), &secretOpts)
	if err != nil {
		c.respondError(ctx, err)
	}
	return client
}

func (c *ApplicationSettings) setKeyvaultSecret(client *azsecrets.Client, secretName, secretValue string, tags map[string]*string) error {
	ctx := context.Background()
	enabled := secretValue != ""

	secretName = strings.Replace(secretName, "_", "-", -1)

	secretParams := azsecrets.SetSecretParameters{
		SecretAttributes: &azsecrets.SecretAttributes{
			Enabled: &enabled,
		},
		Value: &secretValue,
		Tags:  tags,
	}
	_, err := client.SetSecret(ctx, secretName, secretParams, nil)

	return err
}

func (c *ApplicationSettings) getKeyvaultSecret(client *azsecrets.Client, secretName string) (secretValue string) {
	ctx := context.Background()

	secretBundle, err := client.GetSecret(ctx, secretName, "", nil)

	if err == nil {
		secretValue = *secretBundle.Value
	}

	return
}
