package models

type (
	AppConfigPermissions struct {
		AdminGroups []string                  `yaml:"adminGroups"`
		Teams       map[string]*AppConfigTeam `yaml:"teams"`
	}

	AppConfigTeam struct {
		Name    string `yaml:"name"`
		IsAdmin bool   `yaml:"isAdmin"`
		Azure   struct {
			Group            *string `yaml:"group"`
			ServicePrincipal *string `yaml:"servicePrincipal"`
		} `yaml:"azure"`
	}
)
