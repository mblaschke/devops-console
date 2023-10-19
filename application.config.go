package main

import (
	iris "github.com/kataras/iris/v12"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/response"
	"github.com/mblaschke/devops-console/services"
)

type ApplicationConfig struct {
	*Server
}

func NewApplicationConfig(c *Server) *ApplicationConfig {
	app := ApplicationConfig{Server: c}
	return &app
}

func (c *ApplicationConfig) handleApiAppConfig(ctx iris.Context, user *models.User) {
	ret := response.ResponseConfig{}
	ret.User.Username = user.Username

	for _, team := range user.GetTeams() {
		row := response.ResponseConfigTeam{
			Id:   team.Name,
			Name: team.Name,
		}
		ret.Teams = append(ret.Teams, row)
	}

	ret.Quota = map[string]int{
		"team": c.config.Kubernetes.Namespace.Quota.Team,
	}

	// azure
	ret.Azure = response.ResponseConfigAzure{}
	ret.Azure.TenantId = opts.Azure.TenantId
	for _, row := range c.config.Azure.ResourceGroup.Tags {
		tmp := response.ResponseConfigAzureResourceGroupTag{
			Name:        row.Name,
			Label:       row.Label,
			Description: row.Description,
			Type:        row.Type,
			Default:     row.Default,
			Placeholder: row.Placeholder,
		}

		ret.Azure.ResourceGroup.Tags = append(ret.Azure.ResourceGroup.Tags, tmp)
	}

	ret.Azure.RoleAssignment.RoleDefinitions = c.config.Azure.RoleAssignment.RoleDefinitions
	ret.Azure.RoleAssignment.Ttl = c.config.Azure.RoleAssignment.Ttl

	// kubernetes
	ret.Kubernetes = response.ResponseConfigKubernetes{}

	for _, row := range c.config.Kubernetes.Namespace.Settings {
		ret.Kubernetes.Namespace.Settings = append(
			ret.Kubernetes.Namespace.Settings,
			response.ResponseConfigKubernetesNamespaceSetting{
				Name:        row.Name,
				Label:       row.Label,
				Description: row.Description,
				K8sType:     row.K8sType,
				K8sName:     row.K8sName,
				Type:        row.Type,
				Default:     row.Default,
				Placeholder: row.Placeholder,
				Required:    row.Validation.Required,
			},
		)
	}

	for _, row := range c.config.Kubernetes.Namespace.NetworkPolicy {
		ret.Kubernetes.Namespace.NetworkPolicy = append(
			ret.Kubernetes.Namespace.NetworkPolicy,
			response.ResponseConfigKubernetesNamespaceNetworkPolicy{
				Name:        row.Name,
				Description: row.Description,
			},
		)
	}

	ret.Support.Pagerduty.Endpoints = []response.ResponseConfigSupportPagerdutyEndpoint{}
	for key, row := range services.PagerDutyGetEndpointList() {
		ret.Support.Pagerduty.Endpoints = append(
			ret.Support.Pagerduty.Endpoints,
			response.ResponseConfigSupportPagerdutyEndpoint{
				Id:   key,
				Name: row.Name,
			},
		)
	}

	c.responseJson(ctx, ret)
}
