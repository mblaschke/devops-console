package models

type (
	AppConfigSupport struct {
		Pagerduty AppConfigSupportPagerduty `yaml:"pagerduty"`
	}

	AppConfigSupportPagerduty struct {
		AuthToken string                                       `yaml:"authToken"`
		ClientURL string                                       `yaml:"clientURL"`
		Endpoints map[string]AppConfigSupportPagerdutyEndpoint `yaml:"endpoints"`
	}

	AppConfigSupportPagerdutyEndpoint struct {
		RoutingKey string `yaml:"routingKey"`
	}
)
