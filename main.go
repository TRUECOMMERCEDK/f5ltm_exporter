// This is an F5 LTM exporter for getting data from the F5 Local Traffic Management Device
// Author: Thomas Elsgaard <thomas.elsgaard@trucecommerce.com>

package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/truecommercedk/f5ltm_exporter/config"
	"github.com/truecommercedk/f5ltm_exporter/f5"
	"log/slog"
	"maragu.dev/env"
	"net"
	"net/http"
	"os"
	"strconv"
)

var (
	release string

	ltmPoolState = prometheus.NewDesc(
		prometheus.BuildFQName("f5ltm", "", "pool_state"),
		"F5 LTM Pool status",
		[]string{"pool_name", "node_name"}, nil,
	)
)

type Exporter struct {
	config config.Config
}

func NewExporter(config config.Config) *Exporter {

	return &Exporter{
		config: config,
	}

}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	ch <- ltmPoolState
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.UpdateMetrics(ch)
}

func (e *Exporter) UpdateMetrics(ch chan<- prometheus.Metric) {

	f5Api := &f5.Model{
		User: e.config.F5User,
		Pass: e.config.F5Pass,
		Host: e.config.F5Host,
	}

	sessionId, err := f5Api.Authenticate()
	if err != nil {
		slog.Error("Unable to authenticate to F5")
	}

	PoolStats, err := f5Api.GetPoolStats(sessionId)
	if err != nil {
		slog.Error("Unable to retrieve data from f5")
		os.Exit(1)
	}

	for _, v := range PoolStats.Entries {

		switch v.NestedStats.Entries.StatusAvailabilityState.Description {
		case "available":
			ch <- prometheus.MustNewConstMetric(
				ltmPoolState, prometheus.GaugeValue, 1, v.NestedStats.Entries.TmName.Description, e.config.F5Host,
			)
		default:
			ch <- prometheus.MustNewConstMetric(
				ltmPoolState, prometheus.GaugeValue, 0, v.NestedStats.Entries.TmName.Description, e.config.F5Host,
			)

		}

	}
}

func main() {

	_ = env.Load()

	host := env.GetStringOrDefault("HOST", "0.0.0.0")
	port := env.GetIntOrDefault("PORT", 9143)
	metricsPath := env.GetStringOrDefault("METRICS_PATH", "/metrics")

	envConfig := config.Config{

		F5User: env.GetStringOrDefault("F5_USER", ""),
		F5Pass: env.GetStringOrDefault("F5_PASS", ""),
		F5Host: env.GetStringOrDefault("F5_HOST", ""),
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))
	exporter := NewExporter(envConfig)
	prometheus.MustRegister(exporter)
	prometheus.Unregister(collectors.NewGoCollector())

	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>F5 LTM Exporter</title></head>
             <body>
             <h1>F5 LTM Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	slog.Info("F5 Local Traffic Management Device Exporter Started")

	if err := http.ListenAndServe(address, nil); err != nil {
		slog.Error("Error starting server")
	}
}
