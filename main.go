package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	UserAgent = "devops-console/"
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

	if gitTag == "<unknown>" {
		gitTag = fmt.Sprintf("unknown-%v", time.Now().Unix())
	}

	logger, err := logConfig.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
	defer log.Sync() // nolint flushes buffer, if any

	startPrometheus()

	devopsConsole := NewServer(opts.Config.Path)
	startupDuration = time.Since(startupTime)
	devopsConsole.Run(opts.Server.Bind)
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
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
		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())

		srv := &http.Server{
			Addr:         opts.MetricsServer.Bind,
			Handler:      mux,
			ReadTimeout:  opts.MetricsServer.ReadTimeout,
			WriteTimeout: opts.MetricsServer.WriteTimeout,
		}
		log.Fatal(srv.ListenAndServe())
	}()
}
