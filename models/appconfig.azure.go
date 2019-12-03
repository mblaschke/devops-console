package models

type (
	AppConfigAzure struct {
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
