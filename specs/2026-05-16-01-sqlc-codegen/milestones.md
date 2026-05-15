# Milestones: sqlc-codegen

## M1: Extend proto options with ROLE_ENTITY

- **Status:** done
- **Description:** Add `ROLE_ENTITY` to the `Role` enum and update the lint plugin to validate that `ROLE_ENTITY` messages have an `entity` field of type `clarity.plugin.v1.Entity` at field number 1.
- **Acceptance Criteria:**
  - [x] `ROLE_ENTITY = 2` added to `Role` enum in `options.proto`
  - [x] Lint plugin rejects `ROLE_ENTITY` messages missing the `entity` field
  - [x] Lint plugin rejects `ROLE_ENTITY` messages where `entity` is not at field number 1
  - [x] Lint plugin rejects `ROLE_ENTITY` messages where `entity` is not of type `clarity.plugin.v1.Entity`

## M2: Proto-to-PostgreSQL type mapping

- **Status:** pending
- **Description:** Implement the type mapping layer that converts proto field descriptors to PostgreSQL column definitions, covering scalars, well-known types, enums, repeated fields, nested messages, maps, and oneofs.
- **Acceptance Criteria:**
  - [ ] All scalar proto types map to correct PostgreSQL types
  - [ ] Well-known types (`Timestamp`, `Duration`, `Struct`, `Value`) map correctly
  - [ ] Enums map to `TEXT` with `CHECK` constraint listing valid value names
  - [ ] Repeated scalars map to array types
  - [ ] Repeated messages, nested messages, and maps map to `JSONB`
  - [ ] Oneof fields map to nullable columns per variant
  - [ ] Unit tests cover all type mappings

## M3: Schema DDL generation

- **Status:** pending
- **Description:** Generate `schema.sql` with `CREATE SCHEMA` and `CREATE TABLE` statements, inlining `Entity` fields and mapping remaining message fields to columns.
- **Acceptance Criteria:**
  - [ ] Schema name derived as `<provider>_<domain>` from proto package
  - [ ] Table name derived as snake_case of message name
  - [ ] Entity fields inlined as `id UUID PRIMARY KEY`, `created_at TIMESTAMPTZ NOT NULL`, `updated_at TIMESTAMPTZ NOT NULL`
  - [ ] Remaining message fields mapped to columns using M2 type mapping
  - [ ] Multiple `ROLE_ENTITY` messages in a package produce tables in a single `schema.sql`
  - [ ] Output written to `internal/<provider>/<domain>/<version>/sql/schema.sql`

## M4: CRUD query generation

- **Status:** pending
- **Description:** Generate per-entity sqlc-annotated query files with insert, get by id, list, update, and delete operations.
- **Acceptance Criteria:**
  - [ ] Insert query includes all columns, with `id` passed as parameter
  - [ ] Get by id query selects all columns filtered by `id`
  - [ ] List query selects all columns
  - [ ] Update query updates non-primary-key columns filtered by `id`, sets `updated_at`
  - [ ] Delete query removes row by `id`
  - [ ] Each query has correct sqlc annotation (`-- name: <Name> :one`, `:many`, `:exec`, etc.)
  - [ ] One query file per entity at `internal/<provider>/<domain>/<version>/sql/queries/<message_name>.sql`

## M5: sqlc config generation and plugin wiring

- **Status:** pending
- **Description:** Generate `sqlc.yaml` configured for PostgreSQL with pgx, and wire everything into the `protoc-gen-clarity` plugin with output directory support.
- **Acceptance Criteria:**
  - [ ] `sqlc.yaml` generated at `internal/<provider>/<domain>/<version>/sqlc.yaml`
  - [ ] Config references `sql/schema.sql` and `sql/queries/` directory
  - [ ] Config targets PostgreSQL with pgx engine
  - [ ] Plugin accepts output directory flag
  - [ ] End-to-end test: proto input produces valid schema, queries, and sqlc config that passes `sqlc compile`
