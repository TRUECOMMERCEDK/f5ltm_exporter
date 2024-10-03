package prober

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/truecommercedk/f5ltm_exporter/config"
	"github.com/truecommercedk/f5ltm_exporter/f5"
	"log/slog"
	"net/http"
	"regexp"
)

func Handler(w http.ResponseWriter, r *http.Request, c config.Config, logger *slog.Logger) {

	pingCounter := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "pool_state",
			Help:      "F5 LTM Pool status",
			Namespace: "f5ltm",
		},
		[]string{
			"partition_name",
			"pool_name",
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(pingCounter)

	target := r.URL.Query().Get("target")
	if target == "" {
		logger.Error("Target parameter is missing")
		http.Error(w, fmt.Sprintf("Target parameter is missing"), http.StatusBadRequest)
		return
	}

	f5Api := &f5.Model{
		User: c.F5User,
		Pass: c.F5Pass,
		Host: target,
	}

	sessionId, err := f5Api.Authenticate()
	if err != nil {
		logger.Error("Authentication error", slog.Any("err_msg", err))
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}

	PoolStats, err := f5Api.GetPoolStats(sessionId)
	if err != nil {
		logger.Error("Get poolstats error", slog.Any("err_msg", err))
		http.Error(w, fmt.Sprintf("Error:%s", err), http.StatusBadRequest)
		return
	}

	result, _ := regexp.Compile(`/(.*)/(.*)`)

	for _, v := range PoolStats.Entries {

		res := result.FindStringSubmatch(v.NestedStats.Entries.TmName.Description)

		switch v.NestedStats.Entries.StatusAvailabilityState.Description {
		case "available":
			pingCounter.WithLabelValues(res[1], res[2]).Set(1)
		default:
			pingCounter.WithLabelValues(res[1], res[2]).Set(0)
		}

	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
