package prober

import (
	"log/slog"
	"regexp"

	"github.com/elsgaard/f5api"
)

func populatePoolStatsMetrics(metrics *Metrics, data f5api.PoolStats, logger *slog.Logger) {
	re := regexp.MustCompile(`/(.*)/(.*)`)

	for _, entry := range data.Entries {
		tmName := entry.NestedStats.Entries.TmName.Description
		match := re.FindStringSubmatch(tmName)
		if len(match) != 3 {
			logger.Warn("Invalid tmName format", slog.String("tmName", tmName))
			continue
		}
		partition, pool := match[1], match[2]

		state := 0.0
		if entry.NestedStats.Entries.StatusAvailabilityState.Description == "available" {
			state = 1.0
		}

		metrics.PoolState.WithLabelValues(partition, pool).Set(state)
		metrics.ActiveMembers.WithLabelValues(partition, pool).Set(float64(entry.NestedStats.Entries.ActiveMemberCnt.Value))
		metrics.AvailableMembers.WithLabelValues(partition, pool).Set(float64(entry.NestedStats.Entries.AvailableMemberCnt.Value))
		metrics.ConfiguredMembers.WithLabelValues(partition, pool).Set(float64(entry.NestedStats.Entries.MemberCnt.Value))
		metrics.CurrentConnections.WithLabelValues(partition, pool).Set(float64(entry.NestedStats.Entries.ServersideCurConns.Value))
		metrics.TotalConnections.WithLabelValues(partition, pool).Set(float64(entry.NestedStats.Entries.ServersideTotConns.Value))
	}
}

func populateSyncMetrics(metrics *Metrics, status int, logger *slog.Logger) {

	metrics.SyncStatus.Set(float64(status))

}
