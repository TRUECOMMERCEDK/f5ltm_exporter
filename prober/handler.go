package prober

import (
	"f5ltm_exporter/internal/f5api"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler probes the F5 target and exports metrics.
func Handler(w http.ResponseWriter, r *http.Request, f5 *f5api.Model, logger *slog.Logger) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		logger.Error("Missing target parameter")
		return
	}

	// Set F5 target for this scrape
	f5.Host = target

	metrics := createMetrics()

	// --- Collect pool stats ---
	if err := collectPoolStats(f5, metrics, logger); err != nil {
		http.Error(w, "Failed to collect pool stats", http.StatusInternalServerError)
		logger.Error("Failed to collect pool stats",
			slog.String("target", target),
			slog.Any("error", err))
		return
	}

	// --- Collect sync status ---
	if err := collectSyncStatus(f5, metrics, logger); err != nil {
		http.Error(w, "Failed to collect sync status", http.StatusInternalServerError)
		logger.Error("Failed to collect sync status",
			slog.String("target", target),
			slog.Any("error", err))
		return
	}

	logger.Info("F5 scrape successful", slog.String("target", target))
	promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func collectPoolStats(f5 *f5api.Model, metrics *Metrics, logger *slog.Logger) error {
	stats, err := f5.GetPoolStats()
	if err != nil {
		logger.Error("Failed to fetch pool stats", slog.Any("error", err))
		return err
	}
	populatePoolStatsMetrics(metrics, stats, logger)
	return nil
}

func collectSyncStatus(f5 *f5api.Model, metrics *Metrics, logger *slog.Logger) error {
	status, err := f5.GetSyncStatus()
	if err != nil {
		logger.Error("Failed to fetch sync status", slog.Any("error", err))
		return err
	}
	populateSyncMetrics(metrics, status, logger)
	return nil
}
