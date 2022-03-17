package main

import (
	"encoding/json"
)

type (
	ConfigOpts struct {
		// logger
		Logger struct {
			Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
			LogJson bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`
		}

		Kubernetes struct {
			EnableNamespacePodCount bool `long:"enable-namespace-pod-count" env:"ENABLE_NAMESPACE_POD_COUNT" description:"Enable namespace pod count"`
		}

		Azure struct {
			SubscriptionId string `long:"azure.subscriptionid"   env:"AZURE_SUBSCRIPTION_ID"   description:"Azure subscription id"     required:"true"`
			TenantId       string `long:"azure.tenantid"         env:"AZURE_TENANT_ID"         description:"Azure tenant id"           required:"true"`
		}

		// config
		Config struct {
			Path []string `long:"config" env:"CONFIG" description:"Path to config file" default:"./config/default.yaml" env-delim:":"`
			Dump bool     `long:"dump-config" description:"Dump config"`
		}

		// general options
		ServerBind  string `long:"bind"     env:"SERVER_BIND"   description:"Server address"     default:":9000"`
		MetricsBind string `long:"metrics-bind" env:"METRICS_BIND" description:"Server address" default:":9001"`

		Debug                    bool  `long:"debug" description:"Enable debug mode"`
		DisableCsrfProtection    bool  `long:"disable-csrf" env:"DISABLE_CSRF_PROTECTION" description:"Disable CSFR protection"`
		ErrorPunishmentThreshold int64 `long:"error-punishment-threshold" env:"ERROR_PUNISHMENT_THRESHOLD" description:"Error threshold when punishment is executed (logout)" default:"3"`
	}
)

func (o *ConfigOpts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
