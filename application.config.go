package main

import (
	"devops-console/models"
	"devops-console/models/response"
	iris "github.com/kataras/iris/v12"
)

type ApplicationConfig struct {
	*Server
}

func (c *ApplicationConfig) handleApiAppConfig(ctx iris.Context, user *models.User) {
	ret := response.ResponseConfig{}
	ret.User.Username = user.Username

	for _, team := range user.Teams {
		row := response.ResponseConfigTeam{
			Id:   team.Name,
			Name: team.Name,
		}
		ret.Teams = append(ret.Teams, row)
	}

	ret.Quota = map[string]int{
		"team": c.config.Kubernetes.Namespace.Quota.Team,
		"user": c.config.Kubernetes.Namespace.Quota.User,
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

	for _, row := range c.config.Kubernetes.Environments {
		ret.Kubernetes.Environments = append(
			ret.Kubernetes.Environments,
			response.ResponseConfigKubernetesNamespaceEnvironments{
				Environment: row.Name,
				Description: row.Description,
				Template:    row.Template,
			},
		)
	}

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

	// Alertmanager
	ret.Alertmanager.Instances = []string{}
	for _, alertmanagerInstance := range c.config.Alertmanager.Instances {
		ret.Alertmanager.Instances = append(ret.Alertmanager.Instances, alertmanagerInstance.Name)
	}

	c.responseJson(ctx, ret)
}
