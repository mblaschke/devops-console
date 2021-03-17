package models

type (
	AppConfigSettings struct {
		Vault struct {
			Url string
		}

		Kubeconfig map[string]AppConfigKubeconfig

		User []AppConfigSettingItem
		Team []AppConfigSettingItem
	}

	AppConfigKubeconfig struct {
		Name        string
		Description string
		Content     string
	}

	AppConfigSettingItem struct {
		Name        string
		Label       string
		Type        string
		Placeholder string
		Validation  AppInputValidation
		Tags        map[string]string
	}
)
