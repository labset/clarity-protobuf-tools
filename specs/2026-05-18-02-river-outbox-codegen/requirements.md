# Requirements: River Outbox Codegen

> Generate transactional outbox events using River queue for CRUD handlers, enabling async job processing outside the request flow.

## Context

The existing `connect-crud` codegen mode generates Connect RPC handlers that perform CRUD operations directly against Postgres via SQLC. These operations are synchronous and self-contained — there is no mechanism to trigger async processing (e.g. notifications, denormalisation, integrations) in response to data mutations.

We want a new codegen mode (`connect-crud-outbox`) that wraps mutating CRUD operations in a Postgres transaction and atomically enqueues a River job alongside the data change. This guarantees exactly-once delivery semantics for downstream consumers without requiring distributed transactions or polling.

## Requirements

### Functional

- [ ] FR-1: Introduce a new codegen mode `connect-crud-outbox` that composes the existing `sqlc` and `atlas` generators
- [ ] FR-2: Generate handler scaffolding that injects `*river.Client[pgx.Tx]` alongside `*pgxpool.Pool` in the handler deps
- [ ] FR-3: Generate transactional Create/Update/Delete RPC methods that wrap the SQLC call and `river.InsertTx()` in a single Postgres transaction
- [ ] FR-4: Generate Get and List RPC methods unchanged (no outbox, no transaction wrapping)
- [ ] FR-5: Generate outbox event files at `<provider>/<domain>/<version>/outbox/event_<op>_<model>.go` for each mutating operation (create, update, delete)
- [ ] FR-6: Each event struct implements `river.JobArgs` with a `Kind()` method returning `<op>_<model_snake>` (e.g. `create_book`, `update_book`, `delete_book`)
- [ ] FR-7: Create event args contain `EntityID (uuid.UUID)` and `OccurredAt (time.Time)`
- [ ] FR-8: Update event args contain `EntityID (uuid.UUID)`, `FieldMask ([]string)`, and `OccurredAt (time.Time)`
- [ ] FR-9: Delete event args contain `EntityID (uuid.UUID)` and `OccurredAt (time.Time)`
- [ ] FR-10: Mapper and field extraction logic is reused from the existing `connect-crud` generator

### Non-Functional

- [ ] NFR-1: Follow existing codegen conventions — `generator_` file prefix, `embed` for templates, golden file tests
- [ ] NFR-2: Templates live under `internal/codegen/templates/connect-crud-outbox/`
- [ ] NFR-3: Golden test files scoped under `internal/codegen/testdata/golden/connect-crud-outbox/`

### Deferred

- [ ] DFR-1: CDC-based alternative (Postgres logical replication / Debezium) — to be evaluated as a complementary or replacement approach for event emission
- [ ] DFR-2: Actor/principal on event args for audit logging — requires a convention for extracting actor identity from request context

## Constraints

- Must use River's `InsertTx` to guarantee atomicity between the data mutation and the job enqueue
- Event structs are generated but corresponding workers are implemented manually by the consumer
- The `connect-crud` mode remains unchanged — `connect-crud-outbox` is a separate mode

## Out of Scope

- Worker implementation or worker codegen
- Event schema registry or versioning
- Retry policies or dead-letter queue configuration
- River client initialisation or lifecycle management
