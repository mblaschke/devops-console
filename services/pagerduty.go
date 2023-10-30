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

	pagerDutyStaticEndpoints map[string]models.PagerDutyEndpoint
)

func PagerDutySetStaticEndpoints(config models.AppConfig) {
	list := map[string]models.PagerDutyEndpoint{}

	// add static config
	for _, pagerdutyService := range config.Support.Pagerduty.Services {
		key := "__static__:" + pagerdutyService.Name
		list[key] = models.PagerDutyEndpoint{
			Name:       pagerdutyService.Name,
			RoutingKey: pagerdutyService.IntegrationKey,
		}
	}

	pagerDutyStaticEndpoints = list
}

func PagerDutyUpdateEndpointList(ctx context.Context, config models.AppConfig) {
	list := map[string]models.PagerDutyEndpoint{}

	serviceRegexp := regexp.MustCompile(config.Support.Pagerduty.EndpointServiceRegexp)
	integrationRegexp := regexp.MustCompile(config.Support.Pagerduty.EndpointIntegrationRegexp)

	client := pagerduty.NewClient(config.Support.Pagerduty.AuthToken)

	opts := pagerduty.ListServiceOptions{
		Limit:    100,
		Includes: []string{"integrations"},
	}
	response, err := client.ListServicesWithContext(ctx, opts)
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

				key := fmt.Sprintf(`%v:%v`, service.ID, integration.ID)
				name := fmt.Sprintf(config.Support.Pagerduty.EndpointNameTemplate, service.Name, integration.Summary)

				if integration.IntegrationKey != "" {
					list[key] = models.PagerDutyEndpoint{
						Name:       name,
						RoutingKey: integration.IntegrationKey,
					}
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

	list := map[string]models.PagerDutyEndpoint{}

	// add static config
	for key, row := range pagerDutyStaticEndpoints {
		list[key] = row
	}

	// add dynamic config
	for key, row := range pagerDuty.endpointList {
		list[key] = row
	}

	return list
}

func PagerDutySetEndpointList(list map[string]models.PagerDutyEndpoint) {
	pagerDuty.lock.Lock()
	defer pagerDuty.lock.Unlock()

	pagerDuty.endpointList = list
}
