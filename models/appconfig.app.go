package models

import "time"

type (
	Application struct {
		Session struct {
			Type       string
			Expiry     time.Duration
			CookieName string `yaml:"cookieName"`

			Internal struct {
			} `yaml:"internal"`

			SecureCookie struct {
				HashKey  string `yaml:"hashKey"`
				BlockKey string `yaml:"blockKey"`
			} `yaml:"secureCookie"`
		}

		Csrf struct {
			Secret string `yaml:"secret"`
		} `yaml:"csrf"`

		Oauth OAuthConfig `yaml:"oauth"`

		Notification struct {
			Slack struct {
				Webhook string
				Channel string
				Message string
			}
		}

		Kubernetes struct {
			ObjectsPath string `yaml:"objectsPath"`

			ObjectsList map[string]*KubernetesObjectList

			Environments []AppConfigKubernetesEnvironment `yaml:"environments"`

			Namespace KubernetesNamespaceConfig
		} `yaml:"kubernetes"`
	}

	OAuthConfig struct {
		Provider string `yaml:"provider"`

		UsernamePrefix string `yaml:"usernamePrefix"`
		GroupPrefix    string `yaml:"groupPrefix"`

		RedirectUrl string `yaml:"redirectUrl"`

		Azure struct {
			Tenant       string `yaml:"tenant"`
			ClientId     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret"`

			EndpointAuth  string `yaml:"endpointAuth"`
			EndpointToken string `yaml:"endpointToken"`
		} `yaml:"azuread"`

		Filter struct {
			UsernameWhitelist string `yaml:"usernameWhitelist"`
			UsernameBlacklist string `yaml:"usernameBlacklist"`
		} `yaml:"filter"`
	}

	AppConfigKubernetesEnvironment struct {
		Name        string
		Description string
		Template    string
		Quota       string
	}

	KubernetesNamespaceConfig struct {
		Filter struct {
			Access string
			Delete string
			User   string
			Team   string
		}

		Validation struct {
			App  string
			Team string
		}

		Annotations struct {
			Description string
			Immortal    string
		}

		Labels struct {
			Name        string
			User        string
			Team        string
			Environment string
		}

		Role struct {
			Team    string
			User    string
			Private bool
		}

		Quota struct {
			User int
			Team int
		}
	}
)
