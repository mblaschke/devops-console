package models

type (
	SettingsOverall struct {
		Settings AppConfigSettings            `json:"configuration"`
		User     map[string]string            `json:"user"`
		Team     map[string]map[string]string `json:"team"`
	}

	SettingSecret struct {
		TeamName    string `json:"teamName"`
		SettingName string `json:"settingName"`
		Secret      string `json:"-"`
	}
)
