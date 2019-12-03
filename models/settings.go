package models

type (
	SettingsOverall struct {
		Settings AppConfigSettings            `json:"Configuration"`
		User     map[string]string            `json:"User"`
		Team     map[string]map[string]string `json:"Team"`
	}

	SettingSecret struct {
		TeamName    string
		SettingName string
		Secret      string
	}


)
