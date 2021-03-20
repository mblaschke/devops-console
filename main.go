package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

var (
	argparser *flags.Parser
	opts      ConfigOpts
	log       *zap.SugaredLogger

	logConfig = zap.NewProductionConfig()

	PrometheusActions *prometheus.GaugeVec
	startupTime       time.Time
	startupDuration   time.Duration

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

func main() {
	startupTime = time.Now()
	initArgparser()

	logger, err := logConfig.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
	defer log.Sync() // flushes buffer, if any

	startPrometheus()

	devopsConsole := NewServer(opts.Config.Path)
	startupDuration = time.Since(startupTime)
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
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	logConfig.Encoding = "console"

	// verbose level
	if opts.Logger.Verbose {
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	// debug level
	if opts.Debug {
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		logConfig.DisableStacktrace = false
	}

	// json log format
	if opts.Logger.LogJson {
		logConfig.Encoding = "json"
		logConfig.EncoderConfig.TimeKey = ""
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
			panic(err)
		}
	}()
}
