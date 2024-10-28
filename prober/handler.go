package prober

import (
	"f5ltm_exporter/config"
	"f5ltm_exporter/f5"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"regexp"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request, c config.Config, logger *slog.Logger) {

	start := time.Now()

	poolStateGauge := prometheus.NewGaugeVec(
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

	poolStateActiveMemberCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "pool_members_active_total",
			Help:      "F5 LTM Pool active members count",
			Namespace: "f5ltm",
		},
		[]string{
			"partition_name",
			"pool_name",
		},
	)

	poolStateAvailableMemberCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "pool_members_available_total",
			Help:      "F5 LTM Pool available members count",
			Namespace: "f5ltm",
		},
		[]string{
			"partition_name",
			"pool_name",
		},
	)

	poolStateMemberCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "pool_members_configured_total",
			Help:      "F5 LTM Pool configured members count",
			Namespace: "f5ltm",
		},
		[]string{
			"partition_name",
			"pool_name",
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(poolStateGauge)
	registry.MustRegister(poolStateActiveMemberCountGauge)
	registry.MustRegister(poolStateAvailableMemberCountGauge)
	registry.MustRegister(poolStateMemberCountGauge)

	target := r.URL.Query().Get("target")
	if target == "" {
		logger.Error("F5 Device Scrape", slog.String("request_duration_seconds", time.Since(start).String()), slog.String("err_msg", "Target parameter is missing"))
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
		logger.Error("F5 Device Scrape", slog.Float64("request_duration_seconds", time.Since(start).Seconds()), slog.Any("err_msg", err))
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}

	PoolStats, err := f5Api.GetPoolStats(sessionId)
	if err != nil {
		logger.Error("F5 Device Scrape", slog.Float64("request_duration_seconds", time.Since(start).Seconds()), slog.Any("err_msg", err))
		http.Error(w, fmt.Sprintf("Error:%s", err), http.StatusBadRequest)
		return
	}

	result, _ := regexp.Compile(`/(.*)/(.*)`)

	for _, v := range PoolStats.Entries {

		res := result.FindStringSubmatch(v.NestedStats.Entries.TmName.Description)

		switch v.NestedStats.Entries.StatusAvailabilityState.Description {
		case "available":
			poolStateGauge.WithLabelValues(res[1], res[2]).Set(1)
		default:
			poolStateGauge.WithLabelValues(res[1], res[2]).Set(0)
		}

		poolStateActiveMemberCountGauge.WithLabelValues(res[1], res[2]).Set(float64(v.NestedStats.Entries.ActiveMemberCnt.Value))
		poolStateAvailableMemberCountGauge.WithLabelValues(res[1], res[2]).Set(float64(v.NestedStats.Entries.AvailableMemberCnt.Value))
		poolStateMemberCountGauge.WithLabelValues(res[1], res[2]).Set(float64(v.NestedStats.Entries.MemberCnt.Value))
	}

	logger.Info("F5 Device Scrape", slog.Float64("request_duration_seconds", time.Since(start).Seconds()))
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
