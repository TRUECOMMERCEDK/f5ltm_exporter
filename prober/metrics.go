package prober

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	PoolState          *prometheus.GaugeVec
	CurrentConnections *prometheus.GaugeVec
	TotalConnections   *prometheus.GaugeVec
	ActiveMembers      *prometheus.GaugeVec
	AvailableMembers   *prometheus.GaugeVec
	ConfiguredMembers  *prometheus.GaugeVec
	SyncStatus         prometheus.Gauge
	Registry           *prometheus.Registry
}

func createMetrics() *Metrics {
	labels := []string{"partition_name", "pool_name"}

	m := &Metrics{
		PoolState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "pool_state", Help: "F5 LTM Pool status", Namespace: "f5ltm"},
			labels,
		),
		CurrentConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "pool_connections_current", Help: "Current connections", Namespace: "f5ltm"},
			labels,
		),
		TotalConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "pool_connections_total", Help: "Total connections", Namespace: "f5ltm"},
			labels,
		),
		ActiveMembers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "pool_members_active_total", Help: "Active members count", Namespace: "f5ltm"},
			labels,
		),
		AvailableMembers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "pool_members_available_total", Help: "Available members count", Namespace: "f5ltm"},
			labels,
		),
		ConfiguredMembers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "pool_members_configured_total", Help: "Configured members count", Namespace: "f5ltm"},
			labels,
		),
		SyncStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sync_status", Help: "F5 sync status", Namespace: "f5ltm",
		}),
		Registry: prometheus.NewRegistry(),
	}

	m.Registry.MustRegister(
		m.PoolState, m.CurrentConnections, m.TotalConnections,
		m.ActiveMembers, m.AvailableMembers, m.ConfiguredMembers,
		m.SyncStatus,
	)

	return m
}
