package models

type (
	AppConfigSupport struct {
		Pagerduty AppConfigSupportPagerduty `yaml:"pagerduty"`
	}

	AppConfigSupportPagerduty struct {
		AuthToken                 string `yaml:"authToken"`
		ClientURL                 string `yaml:"clientURL"`
		EndpointNameTemplate      string `yaml:"endpointNameTemplate"`
		EndpointServiceRegexp     string `yaml:"endpointServiceRegexp"`
		EndpointIntegrationRegexp string `yaml:"endpointIntegrationRegexp"`
	}

	AppConfigSupportPagerdutyEndpoint struct {
		RoutingKey string `yaml:"routingKey"`
	}
)
