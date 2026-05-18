# Milestones: E2E Test Connect CRUD Outbox

## M1: Module scaffolding and code generation pipeline

- **Status:** pending
- **Description:** Set up the standalone Go module, proto definitions, buf config, and `mise run` tasks so that `buf generate` + `sqlc generate` produces compilable code.
- **Acceptance Criteria:**
  - [ ] `test/connect-crud-outbox/` exists with `go.mod`, `buf.yaml`, `buf.gen.yaml`
  - [ ] Test proto defines a Product entity with all 5 operations using clarity plugin options
  - [ ] `mise run` task orchestrates `buf generate` then `sqlc generate`
  - [ ] Generated code compiles (`go build ./...`)
  - [ ] `.gitignore` excludes generated output

## M2: Testcontainers setup and database bootstrap

- **Status:** pending
- **Description:** Test helper that spins up Postgres via testcontainers, applies River migrations and the generated schema SQL, and provides a connected `pgxpool.Pool` + `river.Client[pgx.Tx]`.
- **Acceptance Criteria:**
  - [ ] Postgres container starts and is reachable
  - [ ] River migrations applied
  - [ ] Generated schema SQL applied
  - [ ] Pool and River client returned to caller

## M3: Connect server and CRUD tests

- **Status:** pending
- **Description:** Wire up the generated handler on a test HTTP server, exercise all 5 CRUD operations via Connect client, and assert correct responses.
- **Acceptance Criteria:**
  - [ ] Create returns entity with generated ID and timestamps
  - [ ] Get retrieves created entity
  - [ ] List returns items with pagination
  - [ ] Update with field mask applies partial changes
  - [ ] Delete soft-deletes (subsequent Get returns not-found)

## M4: Outbox event assertions

- **Status:** pending
- **Description:** Assert that River jobs are correctly enqueued for mutating operations and absent for read operations.
- **Acceptance Criteria:**
  - [ ] Create enqueues `create_product` job with `entity_id` and `occurred_at`
  - [ ] Update enqueues `update_product` job with `entity_id`, `field_mask`, and `occurred_at`
  - [ ] Delete enqueues `delete_product` job with `entity_id` and `occurred_at`
  - [ ] Get and List do not enqueue any jobs
