package formdata

type (
	AzureResourceGroup struct {
		Name     string            `json:"name"`
		Location string            `json:"location"`
		Team     string            `json:"team"`
		Tags     map[string]string `json:"tags"`
	}

	AzureRoleAssignment struct {
		ResourceId     string `json:"resourceId"`
		RoleDefinition string `json:"roleDefinition"`
		Ttl            string `json:"ttl"`
		Reason         string `json:"reason"`
	}
)
