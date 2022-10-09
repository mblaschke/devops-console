package models

import (
	"fmt"
	"time"
)

type (
	Application struct {
		Features map[string]bool

		Session struct {
			Type           string
			Expiry         time.Duration
			CookieName     string `yaml:"cookieName"`
			CookieSecure   bool   `yaml:"cookieSecure"`
			CookieDomain   string `yaml:"cookieDomain"`
			CookieSameSite string `yaml:"cookieSameSite"`

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

		AuthUrl     string `yaml:"authUrl"`
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

	ApplicationFeatures map[string]bool
)

func (a *Application) MainFeatureIsEnabled(main string) bool {
	if v, exists := a.Features[main]; exists && v {
		return true
	}
	return false
}

func (a *Application) FeatureIsEnabled(main, branch string) bool {
	if v, exists := a.Features[main]; !exists || !v {
		return false
	}

	name := fmt.Sprintf("%v-%v", main, branch)
	if v, exists := a.Features[name]; exists && v {
		return true
	}

	return false
}
