package service

import (
	"math"
	"time"
)

// computeDashboardHealthScore computes a 0-100 health score from the metrics returned by the dashboard overview.
//
// Design goals:
// - Backend-owned scoring (UI only displays).
// - Layered scoring: Business Health (70%) + Infrastructure Health (30%)
// - Avoids double-counting (e.g., DB failure affects both infra and business metrics)
// - Conservative + stable: penalize clear degradations; avoid overreacting to missing/idle data.
func computeDashboardHealthScore(now time.Time, overview *OpsDashboardOverview) int {
	if overview == nil {
		return 0
	}

	// Idle/no-data: avoid showing a "bad" score when there is no traffic.
	// UI can still render a gray/idle state based on QPS + error rate.
	if overview.RequestCountSLA <= 0 && overview.RequestCountTotal <= 0 && overview.ErrorCountTotal <= 0 {
		return 100
	}

	businessHealth := computeBusinessHealth(overview)
	infraHealth := computeInfraHealth(now, overview)

	// Weighted combination: 70% business + 30% infrastructure
	score := businessHealth*0.7 + infraHealth*0.3
	return int(math.Round(clampFloat64(score, 0, 100)))
}

// TTFT scoring thresholds, in milliseconds, applied to the P99 time to first token.
//
// The original 1s/3s scale was written for non-reasoning chat completions, where
// the first token follows almost immediately. A reasoning model emits its
// reasoning pass before the first visible token, so several seconds is its
// healthy floor rather than a fault, and any scale that zeroes at 3s reports 0
// permanently and stops carrying information.
//
// Calibrated against this deployment's measured P99 (gpt-5.6-luna via
// ai-plugin.io, 8394 samples over 24h): p50 5.7s, p90 14.3s, p95 21.2s,
// p99 45.5s. Hourly P99 over healthy operation ranged 17s-54s -- a 3x swing
// with no incident behind it, because P99 on a long-tail reasoning workload is
// decided by a handful of slow requests. Thresholds narrower than that band
// would make the score oscillate for no operational reason.
//
// So: full marks at or below 30s, which covers the good end of the observed
// healthy band; zero at 120s, roughly 2.6x the current P99 and far outside
// anything observed while the service was working, where 1% of requests waiting
// two minutes for a first token is genuinely broken. In between it degrades
// linearly, so ordinary drift still moves the number.
//
// These are workload-specific: a deployment serving non-reasoning models should
// scale them back down.
const (
	opsTTFTP99FullScoreMs = 30000.0
	opsTTFTP99ZeroScoreMs = 120000.0
)

// computeBusinessHealth calculates business health score (0-100)
// Components: Error Rate (50%) + TTFT (50%)
func computeBusinessHealth(overview *OpsDashboardOverview) float64 {
	// Error rate score: 1% → 100, 10% → 0 (linear)
	// Combines request errors and upstream errors
	errorScore := 100.0
	errorPct := clampFloat64(overview.ErrorRate*100, 0, 100)
	upstreamPct := clampFloat64(overview.UpstreamErrorRate*100, 0, 100)
	combinedErrorPct := math.Max(errorPct, upstreamPct) // Use worst case
	if combinedErrorPct > 1.0 {
		if combinedErrorPct <= 10.0 {
			errorScore = (10.0 - combinedErrorPct) / 9.0 * 100
		} else {
			errorScore = 0
		}
	}

	// TTFT score: opsTTFTP99FullScoreMs → 100, opsTTFTP99ZeroScoreMs → 0 (linear)
	// Time to first token is critical for user experience
	ttftScore := 100.0
	if overview.TTFT.P99 != nil {
		p99 := float64(*overview.TTFT.P99)
		if p99 > opsTTFTP99FullScoreMs {
			if p99 <= opsTTFTP99ZeroScoreMs {
				ttftScore = (opsTTFTP99ZeroScoreMs - p99) / (opsTTFTP99ZeroScoreMs - opsTTFTP99FullScoreMs) * 100
			} else {
				ttftScore = 0
			}
		}
	}

	// Weighted combination: 50% error rate + 50% TTFT
	return errorScore*0.5 + ttftScore*0.5
}

// computeInfraHealth calculates infrastructure health score (0-100)
// Components: Storage (40%) + Compute Resources (30%) + Background Jobs (30%)
func computeInfraHealth(now time.Time, overview *OpsDashboardOverview) float64 {
	// Storage score: DB critical, Redis less critical
	storageScore := 100.0
	if overview.SystemMetrics != nil {
		if overview.SystemMetrics.DBOK != nil && !*overview.SystemMetrics.DBOK {
			storageScore = 0 // DB failure is critical
		} else if overview.SystemMetrics.RedisOK != nil && !*overview.SystemMetrics.RedisOK {
			storageScore = 50 // Redis failure is degraded but not critical
		}
	}

	// Compute resources score: CPU + Memory
	computeScore := 100.0
	if overview.SystemMetrics != nil {
		cpuScore := 100.0
		if overview.SystemMetrics.CPUUsagePercent != nil {
			cpuPct := clampFloat64(*overview.SystemMetrics.CPUUsagePercent, 0, 100)
			if cpuPct > 80 {
				if cpuPct <= 100 {
					cpuScore = (100 - cpuPct) / 20 * 100
				} else {
					cpuScore = 0
				}
			}
		}

		memScore := 100.0
		if overview.SystemMetrics.MemoryUsagePercent != nil {
			memPct := clampFloat64(*overview.SystemMetrics.MemoryUsagePercent, 0, 100)
			if memPct > 85 {
				if memPct <= 100 {
					memScore = (100 - memPct) / 15 * 100
				} else {
					memScore = 0
				}
			}
		}

		computeScore = (cpuScore + memScore) / 2
	}

	// Background jobs score
	jobScore := 100.0
	failedJobs := 0
	totalJobs := 0
	for _, hb := range overview.JobHeartbeats {
		if hb == nil {
			continue
		}
		totalJobs++
		if hb.LastErrorAt != nil && (hb.LastSuccessAt == nil || hb.LastErrorAt.After(*hb.LastSuccessAt)) {
			failedJobs++
		} else if hb.LastSuccessAt != nil && now.Sub(*hb.LastSuccessAt) > opsJobStalenessTolerance(hb.JobName) {
			failedJobs++
		}
	}
	if totalJobs > 0 && failedJobs > 0 {
		jobScore = (1 - float64(failedJobs)/float64(totalJobs)) * 100
	}

	// Weighted combination
	return storageScore*0.4 + computeScore*0.3 + jobScore*0.3
}

// opsJobDefaultStalenessTolerance is the staleness window applied to jobs that
// run at least every few minutes (metrics collector, alert evaluator, ...).
const opsJobDefaultStalenessTolerance = 15 * time.Minute

// opsJobStalenessTolerance reports how long a background job may go without a
// success before it should count as failed in the infrastructure health score.
//
// A single flat window cannot be correct for every job: ops_cleanup is
// cron-scheduled once a day and ops_preaggregation_daily ticks hourly, so a
// flat 15-minute rule reports both as failed for most of their own period even
// though they never errored. That permanently removed ~4 points from the score
// while hiding whether the frequently-running jobs were actually healthy.
//
// The tolerance is therefore derived from each job's own cadence, with slack
// for a run that starts late or takes a while.
func opsJobStalenessTolerance(jobName string) time.Duration {
	switch jobName {
	case opsCleanupJobName:
		// Cron-scheduled; daily by default and admin-configurable. A daily job
		// that has not succeeded for more than a day is genuinely stuck.
		return 26 * time.Hour
	case opsAggDailyJobName:
		return 2 * opsAggDailyInterval
	case opsAggHourlyJobName:
		return 2 * opsAggHourlyInterval
	default:
		return opsJobDefaultStalenessTolerance
	}
}

func clampFloat64(v float64, min float64, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
