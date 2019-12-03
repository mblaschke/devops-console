package models

type (
	AppConfigSettings struct {
		Vault struct {
			Url string
		}
		User []AppConfigSettingItem
		Team []AppConfigSettingItem
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
