package response

import "devops-console/models"

type (
	ResponseConfig struct {
		User                  ResponseConfigUser
		Teams                 []ResponseConfigTeam
		NamespaceEnvironments []ResponseNamespaceConfig
		Quota                 map[string]int
		Azure                 models.AppConfigAzure
		Kubernetes            models.AppConfigKubernetes
		Alertmanager          struct {
			Instances []string `json:"Instances"`
		} `json:"Alertmanager"`
	}

	ResponseConfigUser struct {
		Name     string
		Username string
	}

	ResponseConfigTeam struct {
		Id   string
		Name string
	}

	ResponseNamespaceConfig struct {
		Environment string
		Description string
		Template    string
	}
)
