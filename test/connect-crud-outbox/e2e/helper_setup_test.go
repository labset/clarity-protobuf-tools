package e2e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

var shared *testEnv

type testEnv struct {
	pool   *pgxpool.Pool
	river  *river.Client[pgx.Tx]
	server *httptest.Server
	client inventoryv1connect.ProductServiceClient
}

func TestMain(m *testing.M) {
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
	if err != nil {
		log.Fatalf("starting postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("getting connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("creating pool: %v", err)
	}

	// Create a dev database for atlas schema normalization
	_, err = pool.Exec(ctx, "CREATE DATABASE atlas_dev")
	if err != nil {
		log.Fatalf("creating atlas_dev database: %v", err)
	}

	// Run River migrations
	riverMigrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		log.Fatalf("creating river migrator: %v", err)
	}
	_, err = riverMigrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		log.Fatalf("running river migrations: %v", err)
	}

	// Apply generated schema via atlas
	devURL, err := replaceDBName(connStr, "atlas_dev")
	if err != nil {
		log.Fatalf("building dev URL: %v", err)
	}

	atlasClient, err := atlasexec.NewClient("../internal/test/inventory/v1", "atlas")
	if err != nil {
		log.Fatalf("creating atlas client: %v", err)
	}

	_, err = atlasClient.SchemaApply(ctx, &atlasexec.SchemaApplyParams{
		URL:         connStr,
		To:          "file://sql/schema.sql",
		DevURL:      devURL,
		Schema:      []string{"test_inventory_v1"},
		AutoApprove: true,
	})
	if err != nil {
		log.Fatalf("applying schema: %v", err)
	}

	// Create River client (no workers — we only inspect enqueued jobs)
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		log.Fatalf("creating river client: %v", err)
	}

	// Start Connect server
	mux := http.NewServeMux()
	path, handler := api.NewProductServiceHandler(api.ProductDeps{
		Pool:  pool,
		River: riverClient,
	})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)

	shared = &testEnv{
		pool:   pool,
		river:  riverClient,
		server: server,
		client: inventoryv1connect.NewProductServiceClient(http.DefaultClient, server.URL),
	}

	code := m.Run()

	server.Close()
	pool.Close()
	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}

func setupTest(t *testing.T) *testEnv {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		_, err := shared.pool.Exec(ctx, "TRUNCATE test_inventory_v1.product")
		require.NoError(t, err)
		_, err = shared.pool.Exec(ctx, "DELETE FROM river_job")
		require.NoError(t, err)
	})
	return shared
}

func replaceDBName(connStr, dbName string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", fmt.Errorf("parsing connection string: %w", err)
	}
	u.Path = "/" + dbName
	return u.String(), nil
}
