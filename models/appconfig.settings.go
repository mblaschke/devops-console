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
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}

	AppConfigSettingItem struct {
		Name        string             `json:"name"`
		Label       string             `json:"label"`
		Type        string             `json:"type"`
		Placeholder string             `json:"placeholder"`
		Validation  AppInputValidation `json:"validation"`
		Tags        map[string]string  `json:"tags"`
	}
)
