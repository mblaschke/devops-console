package formdata

type (
	AzureResourceGroup struct {
		Name     string            `json:"name"`
		Location string            `json:"location"`
		Team     string            `json:"team"`
		Tag      map[string]string `json:"tag"`
	}

	AzureRoleAssignment struct {
		ResourceId     string `json:"resourceId"`
		RoleDefinition string `json:"roleDefinition"`
		Reason         string `json:"reason"`
	}
)
