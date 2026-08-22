//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// A request whose upstream call failed but was recovered by failover is logged in
// ops_error_logs with status_code 200 and is simultaneously counted as a success in
// usage_logs. UpstreamErrorRate divides upstream_excl by requestCountSLA, which is
// built from that success count, so counting the recovered row in the numerator
// reports a failure rate higher than the share of requests that actually failed —
// and because the health score takes max(errorRate, upstreamErrorRate) against a
// hard 10% cliff, that alone could hold the error score at 0 on a healthy service.
//
// upstream_excl must therefore count only client-visible failures, exactly like
// error_sla. The 429/529 counters are deliberately left counting every occurrence:
// they are reported as standalone totals and never divided into a rate.
func TestQueryErrorCounts_UpstreamExclCountsOnlyClientVisibleFailures(t *testing.T) {
	ctx := context.Background()
	_, err := integrationDB.ExecContext(ctx, "TRUNCATE ops_error_logs RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	repo := NewOpsRepository(integrationDB).(*opsRepository)
	now := time.Now()

	insert := func(statusCode int, upstreamStatus int, owner string) {
		_, execErr := integrationDB.ExecContext(ctx, `
			INSERT INTO ops_error_logs (
				error_phase, error_type, severity, status_code,
				upstream_status_code, error_owner, is_business_limited,
				is_count_tokens, created_at
			) VALUES ('upstream', 'upstream_error', 'error', $1, $2, $3, FALSE, FALSE, $4)`,
			statusCode, upstreamStatus, owner, now)
		require.NoError(t, execErr)
	}

	// Recovered by failover: the client got a 200, so this is not a failure.
	insert(200, 500, "provider")
	insert(200, 403, "provider")
	// Genuine client-visible upstream failures.
	insert(502, 500, "provider")
	insert(500, 403, "provider")
	// Rate limits, split into their own counters regardless of recovery.
	insert(200, 429, "provider")
	insert(429, 429, "provider")
	insert(529, 529, "provider")

	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	errorTotal, businessLimited, errorCountSLA, upstreamExcl, upstream429, upstream529, err :=
		repo.queryErrorCounts(ctx, &service.OpsDashboardFilter{}, start, end)
	require.NoError(t, err)

	require.Equal(t, int64(4), errorTotal, "status_code >= 400 rows")
	require.Equal(t, int64(0), businessLimited)
	require.Equal(t, int64(4), errorCountSLA)

	// Only the two non-rate-limited rows the client actually saw fail.
	require.Equal(t, int64(2), upstreamExcl,
		"recovered upstream errors (status_code 200) must not inflate the rate numerator")

	// Occurrence counters keep the recovered rows.
	require.Equal(t, int64(2), upstream429)
	require.Equal(t, int64(1), upstream529)

	// The numerator can never exceed error_sla, so UpstreamErrorRate can never
	// report a failure rate above the real one.
	require.LessOrEqual(t, upstreamExcl, errorCountSLA)
}
