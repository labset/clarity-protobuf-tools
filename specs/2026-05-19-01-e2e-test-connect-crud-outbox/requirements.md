# Requirements: E2E Test Connect CRUD Outbox

> Runtime e2e test using testcontainers to validate that generated connect-crud-outbox code compiles, runs, and behaves correctly against a real Postgres instance with River queue.

## Context

The `connect-crud-outbox` codegen mode is fully implemented with golden file tests that validate template output. However, there is no validation that the generated code actually compiles, runs, and behaves correctly at runtime. A standalone e2e test module with testcontainers will close that gap by exercising the full pipeline: proto definition, code generation, database setup, and runtime CRUD + outbox behavior.

## Requirements

### Functional

- [ ] FR-1: Standalone Go module at `test/connect-crud-outbox/` with its own `go.mod`, proto definitions, `buf.yaml`, and `buf.gen.yaml`
- [ ] FR-2: Test proto defines at least one entity message (e.g. `Product`) with all five operations (create, get, list, update, delete) using clarity plugin options
- [ ] FR-3: `buf generate` produces: protobuf Go code, Connect Go code, and clarity `connect-crud-outbox` output (schema SQL, SQLC config, handler, mapper, RPCs, outbox events)
- [ ] FR-4: SQLC runs against the generated SQL to produce the `db` package
- [ ] FR-5: Testcontainers spins up a Postgres instance for each test run
- [ ] FR-6: River migrations applied to the test database before tests
- [ ] FR-7: Schema SQL (generated) applied to the test database before tests
- [ ] FR-8: Connect server started with real `pgxpool.Pool` + `river.Client[pgx.Tx]`, serving on a test HTTP server
- [ ] FR-9: Tests exercise Create, Get, List, Update (with field mask), and Delete via Connect client
- [ ] FR-10: Tests assert CRUD responses are correct (returned entities, not-found errors, pagination)
- [ ] FR-11: Tests assert River jobs are enqueued in `river_job` table with correct `kind` and `args` for create/update/delete operations
- [ ] FR-12: Tests assert no River jobs enqueued for get/list operations

### Non-Functional

- [ ] NFR-1: Module uses a `mise run` task to orchestrate the generate + SQLC + test pipeline
- [ ] NFR-2: Tests use `testify` assert/require
- [ ] NFR-3: Generated code is gitignored (only protos, buf config, and test files are committed)

### Deferred

- [ ] DFR-1: Transaction atomicity negative test (e.g. force River insert failure and verify data mutation rolled back) — adds complexity, can be added later
- [ ] DFR-2: Ref-type fields in the e2e test entity — start with simple fields, extend later

## Constraints

- Must use testcontainers-go for Postgres lifecycle
- River migrations must be applied before schema SQL (River tables must exist for `InsertTx`)
- Generated code (Go, SQL store) must not be committed — only source protos, buf config, and test files

## Out of Scope

- Worker implementation or execution of River jobs
- Performance or load testing
- CI integration (just local `mise run` for now)
