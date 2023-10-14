package models

type (
	AppConfigAzure struct {
		RoleAssignment struct {
			Filter struct {
				ResourceId AppConfigFilter `yaml:"resourceId"`
			}

			RoleDefinitions []string
			Ttl             []string
		}

		ResourceGroup struct {
			Filter struct {
				Name AppConfigFilter `yaml:"name"`
			}

			Tags []AppConfigAzureResourceGroupTag
		}
	}

	AppConfigAzureResourceGroupTag struct {
		Name           string
		Label          string
		Description    string
		Type           string
		Default        string
		Placeholder    string
		Validation     AppInputValidation
		Transformation AppInputTransformation
	}
)
