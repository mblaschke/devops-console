package main

import (
	"devops-console/models"
	"devops-console/models/response"
	"github.com/dustin/go-humanize"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"runtime"
	"time"
)

type ApplicationGeneral struct {
	*Server
}

func (c *ApplicationGeneral) handleApiAppStats(ctx iris.Context, user *models.User) {
	systemStats := response.NewGeneralStats("System stats")
	systemStats.Add("Golang runtime", runtime.Version())
	systemStats.Add("Architecture", runtime.GOARCH)
	systemStats.Add("OS", runtime.GOOS)
	systemStats.Add("CPUs", humanize.Comma(int64(runtime.NumCPU())))

	systemApp := response.NewGeneralStats("App stats")
	systemApp.Add("Startup time", startupTime.Format(time.RFC1123))
	systemApp.Add("Startup duration", startupDuration.String())
	systemApp.Add("Go routines", humanize.Comma(int64(runtime.NumGoroutine())))

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	systemMemory := response.NewGeneralStats("Memory stats")
	systemMemory.Add("OS reserved", humanize.Bytes(memStats.Sys))
	systemMemory.Add("Heap objects", humanize.Comma(int64(memStats.HeapObjects)))
	systemMemory.Add("Heap allocated", humanize.Bytes(memStats.HeapAlloc))
	systemMemory.Add("Malloc calls", humanize.Comma(int64(memStats.Mallocs)))
	systemMemory.Add("Free calls", humanize.Comma(int64(memStats.Frees)))
	systemMemory.Add("GC runs", humanize.Comma(int64(memStats.NumGC)))
	if memStats.LastGC > 0 {
		systemMemory.Add("GC run", humanize.Time(time.Unix(0, int64(memStats.LastGC))))
	} else {
		systemMemory.Add("GC run", "never")
	}

	ret := []response.GeneralStats{}
	ret = append(ret, *systemStats, *systemApp, *systemMemory)

	PrometheusActions.With(prometheus.Labels{"scope": "general", "type": "stats"}).Inc()

	ctx.JSON(ret)
}

func (c *ApplicationGeneral) handleApiAppConfig(ctx iris.Context, user *models.User) {
	ret := response.ResponseConfig{}
	ret.User.Username = user.Username

	for _, team := range user.Teams {
		row := response.ResponseConfigTeam{
			Id:   team.Name,
			Name: team.Name,
		}
		ret.Teams = append(ret.Teams, row)
	}

	row := response.ResponseConfigTeam{
		Id:   "foobar",
		Name: "foobar",
	}
	ret.Teams = append(ret.Teams, row)


	for _, row := range c.config.App.Kubernetes.Environments {
		tmp := response.ResponseNamespaceConfig{
			Environment: row.Name,
			Description: row.Description,
			Template:    row.Template,
		}

		ret.NamespaceEnvironments = append(ret.NamespaceEnvironments, tmp)
	}

	ret.Quota = map[string]int{
		"team": c.config.App.Kubernetes.Namespace.Quota.Team,
		"user": c.config.App.Kubernetes.Namespace.Quota.User,
	}

	// azure
	ret.Azure = c.config.Azure

	// kubernetes
	ret.Kubernetes = c.config.Kubernetes

	// Alertmanager
	ret.Alertmanager.Instances = []string{}
	for _, alertmanagerInstance := range c.config.Alertmanager.Instances {
		ret.Alertmanager.Instances = append(ret.Alertmanager.Instances, alertmanagerInstance.Name)
	}

	ctx.JSON(ret)
}
