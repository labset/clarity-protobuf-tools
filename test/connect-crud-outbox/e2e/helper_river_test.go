package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type riverJob struct {
	Kind string
	Args json.RawMessage
}

func queryRiverJobs(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []riverJob {
	t.Helper()
	rows, err := pool.Query(ctx, "SELECT kind, args FROM river_job ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var jobs []riverJob
	for rows.Next() {
		var j riverJob
		require.NoError(t, rows.Scan(&j.Kind, &j.Args))
		jobs = append(jobs, j)
	}
	require.NoError(t, rows.Err())
	return jobs
}
