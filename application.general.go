package main

import (
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/response"
)

type ApplicationGeneral struct {
	*Server
}

func NewApplicationGeneral(c *Server) *ApplicationGeneral {
	app := ApplicationGeneral{Server: c}
	return &app
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

	c.responseJson(ctx, ret)
}
