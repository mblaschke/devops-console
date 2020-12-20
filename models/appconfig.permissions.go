package models

type (
	AppConfigPermissions struct {
		Default AppConfigDefault          `yaml:"default"`
		User    map[string]AppConfigUser  `yaml:"user"`
		Group   map[string]AppConfigGroup `yaml:"group"`
		Team    map[string]AppConfigTeam  `yaml:"team"`
	}

	AppConfigDefault struct {
		Teams []string `yaml:"teams"`
	}

	AppConfigUser struct {
		Teams []string `yaml:"teams"`
	}

	AppConfigGroup struct {
		Teams []string `yaml:"teams"`
	}

	AppConfigTeam struct {
		K8sRoleBinding       []TeamK8sPermissions       `yaml:"rolebinding"`
		AzureRoleAssignments []TeamAzureRoleAssignments `yaml:"azureroleassignment"`
		ServiceConnections   []TeamServiceConnections   `yaml:"serviceconnection"`
	}
)
