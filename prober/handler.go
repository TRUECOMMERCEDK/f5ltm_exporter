package prober

import (
	"log/slog"
	"net/http"

	"github.com/elsgaard/f5api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(w http.ResponseWriter, r *http.Request, f5 f5api.Model, logger *slog.Logger) {
	target := r.URL.Query().Get("target")
	f5.Host = target

	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		logger.Error("Missing target parameter")
		return
	}

	metrics := createMetrics()

	token, err := login(f5, logger)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}
	defer logout(f5, token, logger)

	if err := collectPoolStats(f5, token, metrics, logger); err != nil {
		http.Error(w, "Failed to collect stats", http.StatusInternalServerError)
		return
	}

	if err := collectSyncStatus(f5, token, metrics, logger); err != nil {
		http.Error(w, "Failed to collect sync status", http.StatusInternalServerError)
		return
	}

	logger.Info("F5 scrape successful", slog.String("target", target))

	promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

// login logs in and returns the token
func login(f5 f5api.Model, logger *slog.Logger) (string, error) {
	token, err := f5.Login()
	if err != nil {
		logger.Error("Authentication failed", slog.Any("error", err))
		return "", err
	}
	return token, nil
}

// logout logs out but does not crash on failure
func logout(f5 f5api.Model, token string, logger *slog.Logger) {
	if err := f5.Logout(token); err != nil {
		logger.Warn("Logout failed", slog.Any("error", err))
	}
}

// collectPoolStats retrieves and populates metrics
func collectPoolStats(f5 f5api.Model, token string, metrics *Metrics, logger *slog.Logger) error {
	stats, err := f5.GetPoolStats(token)
	if err != nil {
		logger.Error("Failed to fetch pool stats", slog.Any("error", err))
		return err
	}
	populatePoolStatsMetrics(metrics, stats, logger)
	return nil
}

// collectSyncStatus retrieves and populates sync status metrics
func collectSyncStatus(f5 f5api.Model, token string, metrics *Metrics, logger *slog.Logger) error {
	status, err := f5.GetSyncStatus(token)
	if err != nil {
		logger.Error("Failed to fetch sync status", slog.Any("error", err))
		return err
	}
	populateSyncMetrics(metrics, status, logger)
	return nil
}
