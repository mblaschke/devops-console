package models

import (
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
			Message  string
		}
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

)
