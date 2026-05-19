package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/internal/test/inventory/v1/api"
	"github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1/inventoryv1connect"
)

type testEnv struct {
	pool   *pgxpool.Pool
	river  *river.Client[pgx.Tx]
	server *httptest.Server
	client inventoryv1connect.ProductServiceClient
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	// Start Postgres container
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(context.Background()))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Create a dev database for atlas schema normalization
	_, err = pool.Exec(ctx, "CREATE DATABASE atlas_dev")
	require.NoError(t, err)

	// Run River migrations
	riverMigrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	require.NoError(t, err)
	_, err = riverMigrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	require.NoError(t, err)

	// Apply generated schema via atlas
	devURL, err := replaceDBName(connStr, "atlas_dev")
	require.NoError(t, err)

	atlasClient, err := atlasexec.NewClient("../internal/test/inventory/v1", "atlas")
	require.NoError(t, err)

	_, err = atlasClient.SchemaApply(ctx, &atlasexec.SchemaApplyParams{
		URL:         connStr,
		To:          "file://sql/schema.sql",
		DevURL:      devURL,
		Schema:      []string{"test_inventory_v1"},
		AutoApprove: true,
	})
	require.NoError(t, err)

	// Create River client (no workers — we only inspect enqueued jobs)
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	require.NoError(t, err)

	// Start Connect server
	mux := http.NewServeMux()
	path, handler := api.NewProductServiceHandler(api.ProductDeps{
		Pool:  pool,
		River: riverClient,
	})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := inventoryv1connect.NewProductServiceClient(
		http.DefaultClient,
		server.URL,
	)

	return &testEnv{
		pool:   pool,
		river:  riverClient,
		server: server,
		client: client,
	}
}

func replaceDBName(connStr, dbName string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", fmt.Errorf("parsing connection string: %w", err)
	}
	u.Path = "/" + dbName
	return u.String(), nil
}
