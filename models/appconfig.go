package models

import (
	"fmt"
	"github.com/karrick/tparse/v2"
	yaml "gopkg.in/yaml.v2"
	"regexp"
	"strings"
	"time"
)

type (
	AppConfig struct {
		App         Application          `yaml:"application"`
		Settings    AppConfigSettings    `yaml:"settings"`
		Azure       AppConfigAzure       `yaml:"azure"`
		Kubernetes  AppConfigKubernetes  `yaml:"kubernetes"`
		Permissions AppConfigPermissions `yaml:"permissions"`
	}

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

	AppConfigPermissions struct {
		Default AppConfigDefault          `yaml:"default"`
		User    map[string]AppConfigUser  `yaml:"user"`
		Group   map[string]AppConfigGroup `yaml:"group"`
		Team    map[string]AppConfigTeam  `yaml:"team"`
	}

	AppConfigSettings struct {
		Vault struct {
			Url string
		}
		User []AppConfigSettingItem
		Team []AppConfigSettingItem
	}

	AppInputValidation struct {
		Regexp   string
		Required bool
	}

	AppInputTransformation struct {
		Convert string `yaml:"convert"`
	}

	AppConfigSettingItem struct {
		Name        string
		Label       string
		Type        string
		Placeholder string
		Validation  AppInputValidation
		Tags        map[string]string
	}

	AppConfigDefault struct {
		Teams []string `yaml:"teams"`
	}

	AppConfigUser struct {
		Teams []string `yaml:"teams"`
	}

	AppConfigGroup struct {
		Teams []string `yaml:"teams"`
	}

	AppConfigTeam struct {
		K8sRoleBinding       []TeamK8sPermissions       `yaml:"rolebinding"`
		AzureRoleAssignments []TeamAzureRoleAssignments `yaml:"azureroleassignment"`
	}

	AppConfigAzure struct {
		ResourceGroup struct {
			Validation AppInputValidation
			Tags       []AppConfigAzureResourceGroupTag
		}
	}

	AppConfigAzureResourceGroupTag struct {
		Name           string
		Label          string
		Description    string
		Type           string
		Default        string
		Placeholder    string
		Validation     AppInputValidation
		Transformation AppInputTransformation
	}

	AppConfigNamespaceSettings struct {
		Name           string
		Label          string
		Description    string
		K8sType        string
		K8sName        string
		Type           string
		Default        string
		Placeholder    string
		Validation     AppInputValidation
		Transformation AppInputTransformation
	}

	AppConfigKubernetes struct {
		Namespace struct {
			Settings []AppConfigNamespaceSettings
		}
	}

	AppConfigKubernetesEnvironment struct {
		Name        string
		Description string
		Template    string
		Quota       string
	}
)

func AppConfigCreateFromYaml(yamlString string) (c *AppConfig, err error) {
	err = yaml.Unmarshal([]byte(yamlString), &c)
	return
}

func (v *AppInputValidation) HumanizeString() (ret string) {
	validationList := []string{}

	if v.Regexp != "" {
		validationList = append(validationList, fmt.Sprintf("regexp:%v", v.Regexp))
	}

	if v.Required {
		validationList = append(validationList, "required")
	}

	if len(validationList) >= 1 {
		ret = strings.Join(validationList, "; ")
	}

	return
}

func (v *AppInputValidation) Validate(value string) (status bool) {
	status = false

	if value == "" && !v.Required {
		return true
	}

	if v.Regexp != "" {
		validationRegexp := regexp.MustCompile(v.Regexp)

		if validationRegexp.MatchString(value) {
			status = true
		}
	} else {
		status = true
	}

	return
}

func (v *AppInputTransformation) Transform(value string) (ret *string) {
	value = strings.TrimSpace(value)

	// skip empty values
	if value == "" {
		ret = &value
		return
	}

	switch v.Convert {
	case "timestamp":
		// check if relative duration
		if timestamp, err := tparse.AddDuration(time.Now(), value); err == nil {
			timestamp := timestamp.Format(time.RFC3339)
			ret = &timestamp
			break
		}

		// check if timestamp
		timeFormats := []string{
			// prefered format
			time.RFC3339,

			// human format
			"2006-01-02 15:04:05 +07:00",
			"2006-01-02 15:04:05 MST",
			"2006-01-02 15:04:05",
			"2006-01-02",

			// allowed formats
			time.RFC822,
			time.RFC822Z,
			time.RFC850,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC3339Nano,
		}

		for _, timeFormat := range timeFormats {
			if timestamp, err := time.Parse(timeFormat, value); err == nil && timestamp.Unix() > 0 {
				timestamp := timestamp.Format(time.RFC3339)
				ret = &timestamp
				break
			}
		}

		break
	case "":
		ret = &value
		break
	}

	return
}
