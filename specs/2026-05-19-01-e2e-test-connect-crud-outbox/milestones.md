# Milestones: E2E Test Connect CRUD Outbox

## M1: Module scaffolding and code generation pipeline

- **Status:** done
- **Description:** Set up the standalone Go module, proto definitions, buf config, and `mise run` tasks so that `buf generate` + `sqlc generate` produces compilable code.
- **Acceptance Criteria:**
  - [x] `test/connect-crud-outbox/` exists with `go.mod`, `buf.yaml`, `buf.gen.yaml`
  - [x] Test proto defines a Product entity with all 5 operations using clarity plugin options
  - [x] `mise run` task orchestrates `buf generate` then `sqlc generate`
  - [x] Generated code compiles (`go build ./...`)
  - [x] `.gitignore` excludes generated output

## M2: Testcontainers setup and database bootstrap

- **Status:** done
- **Description:** Test helper that spins up Postgres via testcontainers, applies River migrations and the generated schema SQL, and provides a connected `pgxpool.Pool` + `river.Client[pgx.Tx]`.
- **Acceptance Criteria:**
  - [x] Postgres container starts and is reachable
  - [x] River migrations applied
  - [x] Generated schema SQL applied
  - [x] Pool and River client returned to caller

## M3: Connect server and CRUD tests

- **Status:** done
- **Description:** Wire up the generated handler on a test HTTP server, exercise all 5 CRUD operations via Connect client, and assert correct responses.
- **Acceptance Criteria:**
  - [x] Create returns entity with generated ID and timestamps
  - [x] Get retrieves created entity
  - [x] List returns items with pagination
  - [x] Update with field mask applies partial changes
  - [x] Delete soft-deletes (subsequent Get returns not-found)

## M4: Outbox event assertions

- **Status:** done
- **Description:** Assert that River jobs are correctly enqueued for mutating operations and absent for read operations.
- **Acceptance Criteria:**
  - [x] Create enqueues `create_product` job with `entity_id` and `occurred_at`
  - [x] Update enqueues `update_product` job with `entity_id`, `field_mask`, and `occurred_at`
  - [x] Delete enqueues `delete_product` job with `entity_id` and `occurred_at`
  - [x] Get and List do not enqueue any jobs
