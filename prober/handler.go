package prober

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/truecommercedk/f5ltm_exporter/config"
	"github.com/truecommercedk/f5ltm_exporter/f5"
	"log/slog"
	"net/http"
	"regexp"
)

func Handler(w http.ResponseWriter, r *http.Request, c config.Config) {

	pingCounter := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "pool_state",
			Help:      "F5 LTM Pool status",
			Namespace: "f5ltm",
		},
		[]string{
			"node_name",
			"partition_name",
			"pool_name",
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(pingCounter)

	target := r.URL.Query().Get("target")
	if target != "" {

		f5Api := &f5.Model{
			User: c.F5User,
			Pass: c.F5Pass,
			Host: target,
		}

		sessionId, err := f5Api.Authenticate()
		if err != nil {
			slog.Error("Unable to authenticate to F5")
		}

		PoolStats, err := f5Api.GetPoolStats(sessionId)
		if err != nil {
			slog.Error("Unable to retrieve data from f5")
			//os.Exit(1)
		}

		r, _ := regexp.Compile(`/(.*)/(.*)`)

		for _, v := range PoolStats.Entries {

			res := r.FindStringSubmatch(v.NestedStats.Entries.TmName.Description)

			switch v.NestedStats.Entries.StatusAvailabilityState.Description {
			case "available":
				pingCounter.WithLabelValues(target, res[1], res[2]).Set(1)
			default:
				pingCounter.WithLabelValues(target, res[1], res[2]).Set(0)
			}

		}

	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
