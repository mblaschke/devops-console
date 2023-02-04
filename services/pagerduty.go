package services

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/PagerDuty/go-pagerduty"

	"github.com/mblaschke/devops-console/models"
)

var (
	pagerDuty = struct {
		endpointList map[string]models.PagerDutyEndpoint
		lock         sync.RWMutex
	}{
		endpointList: map[string]models.PagerDutyEndpoint{},
		lock:         sync.RWMutex{},
	}
)

func PagerDutyUpdateEndpointList(ctx context.Context, config models.AppConfig) {
	list := map[string]models.PagerDutyEndpoint{}

	serviceRegexp := regexp.MustCompile(config.Support.Pagerduty.EndpointServiceRegexp)
	integrationRegexp := regexp.MustCompile(config.Support.Pagerduty.EndpointIntegrationRegexp)

	client := pagerduty.NewClient(config.Support.Pagerduty.AuthToken)

	opts := pagerduty.ListServiceOptions{
		Limit: 100,
	}
	response, err := client.ListServicesWithContext(ctx, pagerduty.ListServiceOptions{})
	if err != nil {
		panic(err)
	}

	for {
		for _, service := range response.Services {
			if !serviceRegexp.MatchString(service.Name) {
				continue
			}

			for _, integration := range service.Integrations {
				if !integrationRegexp.MatchString(integration.Summary) {
					continue
				}

				integrationDetail, err := client.GetIntegrationWithContext(ctx, service.ID, integration.ID, pagerduty.GetIntegrationOptions{})
				if err != nil {
					panic(err)
				}

				key := fmt.Sprintf(`%v:%v`, service.ID, integration.ID)
				name := fmt.Sprintf(config.Support.Pagerduty.EndpointNameTemplate, service.Name, integration.Summary)

				list[key] = models.PagerDutyEndpoint{
					Name:       name,
					RoutingKey: integrationDetail.IntegrationKey,
				}
			}
		}

		if response.More {
			opts.Offset = response.Offset + response.Limit
			response, err = client.ListServicesWithContext(ctx, opts)
			if err != nil {
				panic(err)
			}
		} else {
			break
		}
	}

	PagerDutySetEndpointList(list)
}

func PagerDutyGetEndpointList() map[string]models.PagerDutyEndpoint {
	pagerDuty.lock.RLock()
	defer pagerDuty.lock.RUnlock()

	return pagerDuty.endpointList
}

func PagerDutySetEndpointList(list map[string]models.PagerDutyEndpoint) {
	pagerDuty.lock.Lock()
	defer pagerDuty.lock.Unlock()

	pagerDuty.endpointList = list
}
