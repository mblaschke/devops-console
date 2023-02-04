package models

type (
	PagerDutyEndpoint struct {
		Name       string `json:"name"`
		RoutingKey string `json:"routingKey"`
	}
)
