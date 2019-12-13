package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)

const (
	Version = "2.2.2"
)

var (
	argparser *flags.Parser
	args      []string

	PrometheusActions *prometheus.GaugeVec
	startupTime       time.Time
	startupDuration   time.Duration
)

var opts struct {
	Config      []string `long:"config" env:"CONFIG" description:"Path to config file" default:"./config/default.yaml" env-delim:":"`
	ServerBind  string   `long:"server-bind" env:"SERVER_BIND" description:"Server address" default:":9000"`
	MetricsBind string   `long:"metrics-bind" env:"METRICS_BIND" description:"Server address" default:":9001"`
	DumpConfig  bool     `long:"dump-config" description:"Dump config"`
	Debug       bool     `long:"debug" description:"Enable debug mode"`
	DisableCsrfProtection bool `long:"disable-csrf" env:"DISABLE_CSRF_PROTECTION" description:"Disable CSFR protection"`
	ErrorPunishmentThreshold int64 `long:"error-punishment-threshold" env:"ERROR_PUNISHMENT_THRESHOLD" description:"Error threshold when punishment is executed (logout)" default:"5"`
}

func main() {
	startupTime = time.Now()
	initArgparser()
	startPrometheus()

	devopsConsole := NewServer(opts.Config)
	startupDuration = time.Now().Sub(startupTime)
	devopsConsole.Run(opts.ServerBind)
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println(err)
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

func startPrometheus() {
	PrometheusActions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "devopsconsole_actions",
			Help: "DevOps Console actions",
		},
		[]string{
			"scope",
			"type",
		},
	)
	prometheus.MustRegister(PrometheusActions)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(opts.MetricsBind, nil); err != nil {
			fmt.Println(fmt.Sprintf("ERROR: %s", err))
		}
	}()
}
