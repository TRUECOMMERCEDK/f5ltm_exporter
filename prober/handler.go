package prober

import (
	"f5ltm_exporter/config"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(w http.ResponseWriter, r *http.Request, cfg config.Config, logger *slog.Logger) {
	start := time.Now()

	target := r.URL.Query().Get("target")
	if target == "" {
		logger.Error("Missing target parameter",
			slog.String("duration", time.Since(start).String()))
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	metrics := createMetrics()

	// Get data from F5
	data, err := getF5Stats(target, cfg)
	if err != nil {
		logger.Error("F5 stats retrieval failed",
			slog.Float64("duration_seconds", time.Since(start).Seconds()),
			slog.Any("error", err),
		)
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}
	defer data.Client.Logout(data.SessionID)

	// Populate metrics
	populateMetrics(metrics, data, logger)

	logger.Info("F5 scrape successful",
		slog.Float64("duration_seconds", time.Since(start).Seconds()),
	)

	// Serve metrics
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, r)
}
