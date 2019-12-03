package models

type (

	AppConfigKubernetes struct {
		Namespace struct {
			Settings []AppConfigNamespaceSettings
		}
	}

	AppConfigNamespaceSettings struct {
		Name           string
		Label          string
		Description    string
		K8sType        string
		K8sName        string
		Type           string
		Default        string
		Placeholder    string
		Validation     AppInputValidation
		Transformation AppInputTransformation
	}
)
