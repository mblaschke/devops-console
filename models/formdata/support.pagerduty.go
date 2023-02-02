package formdata

import (
	"errors"
)

type (
	SupportPagerduty struct {
		Team          string `yaml:"team"`
		Endpoint      string `yaml:"endpoint"`
		ResourceType  string `yaml:"resourceType"`
		Location      string `yaml:"location"`
		ResourceGroup string `yaml:"resourceGroup"`
		Resource      string `yaml:"resource"`
		Message       string `yaml:"message"`
		Contact       string `yaml:"contact"`
	}
)

func (a *SupportPagerduty) Validate() (ret *SupportPagerduty, err error) {
	ret = a
	err = nil

	if ret.Team == "" {
		return nil, errors.New("invalid or empty team")
	}

	if ret.Endpoint == "" {
		return nil, errors.New("invalid or empty endpoint")
	}

	if ret.ResourceType == "" {
		return nil, errors.New("invalid or empty ResourceType")
	}

	if ret.Location == "" {
		return nil, errors.New("invalid or empty Location")
	}

	if ret.ResourceGroup == "" {
		return nil, errors.New("invalid or empty ResourceGroup")
	}

	if ret.Resource == "" {
		return nil, errors.New("invalid or empty Resource")
	}

	if ret.Message == "" {
		return nil, errors.New("invalid or empty Message")
	}

	if ret.Contact == "" {
		return nil, errors.New("invalid or empty Contact")
	}

	return a, nil
}
