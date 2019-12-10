package models

import (
	"context"
	"errors"
	"fmt"
	alertmanager "github.com/prometheus/alertmanager/api/v2/client"
	httptransport "github.com/go-openapi/runtime/client"
	"net"
	"net/url"
	"time"
)


type (
	AppConfigAlertmanager struct {
		Instances []AppConfigAlertmanagerInstance `yaml:"instances"`
	}

	AppConfigAlertmanagerInstance struct {
		Name string  `yaml:"name"`
		Url string  `yaml:"url"`
		Auth *struct {
			Type string  `yaml:"type"`
			Username string  `yaml:"username"`
			Password string  `yaml:"password"`
		} `yaml:"auth"`
	}
)

func (a *AppConfigAlertmanager) GetAlertmanagerInstance(name string) (*alertmanager.Alertmanager, error) {
	var config *AppConfigAlertmanagerInstance

	for _, row := range a.Instances {
		if row.Name == name {
			config = &row
			break
		}
	}

	if config == nil {
		return nil, errors.New("Invalid alertmanager instance")
	}

	configUrl, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	hostScheme := configUrl.Scheme
	hostName := configUrl.Hostname()
	hostPort := configUrl.Port()

	if hostPort == "" {
		switch (hostScheme) {
		case "http":
			hostPort = "80"
			break;
		case "https":
			hostPort = "443"
			break;
		}
	}

	transport := httptransport.New(net.JoinHostPort(hostName, hostPort), fmt.Sprintf("%v/api/v2/", configUrl.Path), []string{hostScheme})
	context, _ := context.WithTimeout(context.Background(), time.Duration(20 * time.Second))
	transport.Context = context

	if config.Auth != nil {
		switch (config.Auth.Type) {
		case "basic":
			transport.DefaultAuthentication = httptransport.BasicAuth(config.Auth.Username, config.Auth.Password)
			break;
		default:
			return nil, errors.New("Invalid authentication")
		}
	}

	client := alertmanager.New(transport, nil)

	return client, nil
}
