package formdata

import (
	"errors"
	"fmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	"strings"
	"time"
)

type (
	AlertmanagerForm struct {
		Team    string         `json:"team"`
		Silence models.Silence `json:"silence"`
	}
)

func (a *AlertmanagerForm) Validate() (ret *AlertmanagerForm, err error) {
	ret = a
	err = nil

	if ret.Team == "" {
		return nil, errors.New("invalid or empty team")
	}

	if ret.Silence.Comment == nil {
		return nil, errors.New("invalid or empty comment")
	}

	if ret.Silence.StartsAt == nil {
		return nil, errors.New("invalid or empty startsAt")
	}

	if ret.Silence.EndsAt == nil {
		return nil, errors.New("invalid or empty endsAt")
	}

	// check if time range is correct
	startsAt := time.Time(*ret.Silence.StartsAt)
	endsAt := time.Time(*ret.Silence.EndsAt)
	if startsAt.After(endsAt) {
		return nil, errors.New("startsAt must be before endsAt")
	}

	matcherList := models.Matchers{}
	for _, matcher := range a.Silence.Matchers {
		if (matcher.Name == nil || *matcher.Name == "") && (matcher.Value == nil || *matcher.Value == "") {
			continue
		}

		if matcher.Name == nil || *matcher.Name == "" {
			return nil, errors.New("match label cannot be empty")
		}

		if matcher.Value == nil || *matcher.Value == "" {
			return nil, errors.New("match value cannot be empty")
		}

		if matcher.IsRegex == nil {
			isRegex := false
			matcher.IsRegex = &isRegex
		}

		matcherList = append(matcherList, matcher)
	}
	ret.Silence.Matchers = matcherList

	if len(ret.Silence.Matchers) == 0 {
		return nil, errors.New("at least one matcher needed")
	}

	return
}

func (a *AlertmanagerForm) ToString(id string) (ret string) {
	parts := []string{}

	commentParts := strings.SplitN(*a.Silence.Comment, "\n", 2)
	title := strings.TrimSpace(commentParts[0])
	parts = append(parts, fmt.Sprintf("\"%v\" [id:%v]", title, id))

	parts = append(parts, fmt.Sprintf("startsAt:%v", time.Time(*a.Silence.StartsAt).String()))
	parts = append(parts, fmt.Sprintf("endsAt:%v", time.Time(*a.Silence.EndsAt).String()))

	return strings.Join(parts, " ")
}

func (a *AlertmanagerForm) ToMarkdown(id string) *string {
	parts := []string{}

	commentParts := strings.SplitN(*a.Silence.Comment, "\n", 2)
	title := strings.TrimSpace(commentParts[0])
	parts = append(parts, fmt.Sprintf("id: %v", id))
	parts = append(parts, fmt.Sprintf("title: %v", title))
	parts = append(parts, fmt.Sprintf("created by: %v", *a.Silence.CreatedBy))
	parts = append(parts, fmt.Sprintf("startsAt: %v", time.Time(*a.Silence.StartsAt).String()))
	parts = append(parts, fmt.Sprintf("endsAt: %v", time.Time(*a.Silence.EndsAt).String()))

	ret := strings.Join(parts, "\n")
	return &ret
}
