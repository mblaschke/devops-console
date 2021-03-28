package models

type (
	AppConfigAzure struct {
		RoleAssignment struct {
			RoleDefinitions []string
			Ttl []string
		}

		ResourceGroup struct {
			Validation AppInputValidation
			Tags       []AppConfigAzureResourceGroupTag
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
