package models

import (
	networkingV1 "k8s.io/api/networking/v1"
	"time"
)

type (
	Application struct {
		Session struct {
			Type         string
			Expiry       time.Duration
			CookieName   string `yaml:"cookieName"`
			CookieSecure bool   `yaml:"cookieSecure"`
			CookieDomain string `yaml:"cookieDomain"`

			Internal struct {
			} `yaml:"internal"`

			SecureCookie struct {
				HashKey  string `yaml:"hashKey"`
				BlockKey string `yaml:"blockKey"`
			} `yaml:"secureCookie"`

			Redis struct {
				Addr      string        `yaml:"addr"`
				Timeout   time.Duration `yaml:"timeout"`
				MaxActive int           `yaml:"maxActive"`
				Password  string        `yaml:"password"`
				Database  string        `yaml:"database"`
				Prefix    string        `yaml:"prefix"`
				Delim     string        `yaml:"delim"`
			} `yaml:"redis"`
		}

		Csrf struct {
			Secret string `yaml:"secret"`
		} `yaml:"csrf"`

		Oauth OAuthConfig `yaml:"oauth"`

		Notification struct {
			Channels []string
			Message string
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
			Description   string
			Immortal      string
			NetworkPolicy string `yaml:"networkPolicy"`
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

		NetworkPolicy []ApplicationKubernetesNetworkPolicy `yaml:"networkPolicy"`
	}

	ApplicationKubernetesNetworkPolicy struct {
		Name        string
		Description string
		Path        string
		netpol      *networkingV1.NetworkPolicy
	}
)

func (netpol *ApplicationKubernetesNetworkPolicy) SetKubernetesObject(obj *networkingV1.NetworkPolicy) {
	netpol.netpol = obj
}

func (netpol *ApplicationKubernetesNetworkPolicy) GetKubernetesObject() *networkingV1.NetworkPolicy {
	return netpol.netpol
}
