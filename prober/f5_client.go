package prober

import (
	"f5ltm_exporter/config"
	"f5ltm_exporter/f5"
	"fmt"
)

type F5Data struct {
	PoolStats  *f5.PoolStats
	SyncStatus float64
	SessionID  string
	Client     *f5.Model
}

func getF5Stats(target string, cfg config.Config) (*F5Data, error) {
	client := &f5.Model{
		User: cfg.F5User,
		Pass: cfg.F5Pass,
		Host: target,
	}

	sessionID, err := client.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	poolStats, err := client.GetPoolStats(sessionID)
	if err != nil {
		return nil, fmt.Errorf("pool stats fetch failed: %w", err)
	}

	syncStatus, err := client.GetSyncStatus(sessionID)
	if err != nil {
		return nil, fmt.Errorf("sync status fetch failed: %w", err)
	}

	return &F5Data{
		PoolStats:  &poolStats,
		SyncStatus: float64(syncStatus),
		SessionID:  sessionID,
		Client:     client,
	}, nil
}
