package models

import (
	"errors"
	"regexp"
)

type (
	AppConfigFilter struct {
		Allow string
		allow *regexp.Regexp
		Deny  string
		deny  *regexp.Regexp
	}

	appConfigFilterInternal struct {
		Allow string
		Deny  string
	}
)

func (f *AppConfigFilter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	f.Allow = ""
	f.allow = nil
	f.Deny = ""
	f.deny = nil

	complexVal := appConfigFilterInternal{}
	if complexErr := unmarshal(&complexVal); complexErr == nil {
		f.Allow = complexVal.Allow
		f.Deny = complexVal.Deny
	} else {
		singleVal := ""
		if singleErr := unmarshal(&singleVal); singleErr == nil {
			f.Allow = singleVal
		} else {
			return errors.Join(complexErr, singleErr)
		}
	}

	if f.Allow != "" {
		f.allow = regexp.MustCompile(f.Allow)
	}

	if f.Deny != "" {
		f.deny = regexp.MustCompile(f.Deny)
	}
	return nil
}

func (f *AppConfigFilter) Validate(val string) bool {
	if f.allow != nil && !f.allow.MatchString(val) {
		return false
	}

	if f.deny != nil && f.deny.MatchString(val) {
		return false
	}

	return true
}
