package models

type (
	AppConfigSupport struct {
		Pagerduty AppConfigSupportPagerduty `yaml:"pagerduty"`
	}

	AppConfigSupportPagerduty struct {
		AuthToken      string `yaml:"authToken"`
		IntegrationKey string `yaml:"integrationKey"`
	}
)
