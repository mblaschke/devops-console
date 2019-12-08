package formdata

import (
	"errors"
	"github.com/prometheus/alertmanager/api/v2/models"
)

type (
	AlertmanagerForm struct {
		Team string
		Silence models.Silence
	}
)

func (a *AlertmanagerForm) Validate() (ret *AlertmanagerForm, err error) {
	ret = a
	err = nil

	if ret.Team == "" {
		return nil, errors.New("Invalid or empty team")
	}

	if ret.Silence.Comment == nil {
		return nil, errors.New("Invalid or empty comment")
	}

	if ret.Silence.StartsAt == nil {
		return nil, errors.New("Invalid or empty startsAt")
	}

	if ret.Silence.EndsAt == nil {
		return nil, errors.New("Invalid or empty endsAt")
	}

	matcherList := models.Matchers{}
	for _, matcher := range a.Silence.Matchers {
		if (matcher.Name == nil || *matcher.Name == "" ) && (matcher.Value == nil || *matcher.Value == "") {
			continue
		}

		if matcher.Name == nil || *matcher.Name == "" {
			return nil, errors.New("Matche label cannot be empty")
		}

		if matcher.Value == nil || *matcher.Value == "" {
			return nil, errors.New("Match value cannot be empty")
		}

		if matcher.IsRegex == nil {
			isRegex := false
			matcher.IsRegex = &isRegex
		}

		matcherList = append(matcherList, matcher)
	}
	ret.Silence.Matchers = matcherList

	if len(ret.Silence.Matchers) == 0 {
		return nil, errors.New("At least one matcher needed")
	}

	// validate by alertmanager (client-side)
	err = ret.Silence.Validate(nil)
	if err != nil {
		return nil, err
	}


	return
}
