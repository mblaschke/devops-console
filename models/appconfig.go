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
		App          Application           `yaml:"application"`
		Settings     AppConfigSettings     `yaml:"settings"`
		Azure        AppConfigAzure        `yaml:"azure"`
		Kubernetes   AppConfigKubernetes   `yaml:"kubernetes"`
		Permissions  AppConfigPermissions  `yaml:"permissions"`
		Alertmanager AppConfigAlertmanager `yaml:"alertmanager"`
	}

	AppInputValidation struct {
		Regexp   string
		Required bool
	}

	AppInputTransformation struct {
		Convert string `yaml:"convert"`
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

	case "":
		ret = &value
	}

	return
}
