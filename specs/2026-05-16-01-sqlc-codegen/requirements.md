# Requirements: sqlc-codegen

> Generate PostgreSQL schema DDL and sqlc query files from proto messages annotated with clarity message options.

## Context

The `clarity-protobuf-tools` project provides a protoc plugin (`protoc-gen-clarity`) and a lint plugin (`clarity-lint-plugin`) for working with annotated proto messages. This spec covers extending the toolchain to generate database schema and query files suitable for use with sqlc, targeting PostgreSQL with the pgx driver.

Proto messages annotated with `ROLE_ENTITY` must include a field `entity` of type `clarity.plugin.v1.Entity` at index 1. The codegen plugin inlines the base entity fields (`id`, `created_at`, `updated_at`) into the generated table and maps the remaining message fields to columns.

## Requirements

### Functional

- [x] FR-1: Add `ROLE_ENTITY = 2` to the `Role` enum in `clarity/plugin/v1/options.proto`
- [x] FR-2: The `clarity-lint-plugin` validates that messages with `ROLE_ENTITY` have a field named `entity` of type `clarity.plugin.v1.Entity` at field number 1
- [x] FR-3: The `protoc-gen-clarity` plugin accepts an output directory flag and emits files to it
- [x] FR-4: Generate `internal/<provider>/<domain>/<version>/sql/schema.sql` containing `CREATE SCHEMA` and `CREATE TABLE` statements for all `ROLE_ENTITY` messages in a package
- [x] FR-5: Generate `internal/<provider>/<domain>/<version>/sql/queries/<message_name>.sql` with sqlc-annotated CRUD queries (insert, get by id, list, update, delete) per entity
- [x] FR-6: Generate `internal/<provider>/<domain>/<version>/sqlc.yaml` configured for PostgreSQL with pgx, referencing the schema and queries directory
- [x] FR-7: Derive the PostgreSQL schema name as `<provider>_<domain>` from proto package `<provider>.<domain>.<version>`
- [x] FR-8: Derive the table name as snake_case of the proto message name
- [x] FR-9: Inline `clarity.plugin.v1.Entity` fields as columns: `id UUID PRIMARY KEY`, `created_at TIMESTAMPTZ NOT NULL`, `updated_at TIMESTAMPTZ NOT NULL`
- [x] FR-10: Map proto scalar types to PostgreSQL types:
  - `string` -> `TEXT`
  - `bytes` -> `BYTEA`
  - `bool` -> `BOOLEAN`
  - `int32`, `sint32`, `sfixed32` -> `INTEGER`
  - `int64`, `sint64`, `sfixed64` -> `BIGINT`
  - `uint32`, `fixed32` -> `INTEGER`
  - `uint64`, `fixed64` -> `BIGINT`
  - `float` -> `REAL`
  - `double` -> `DOUBLE PRECISION`
- [x] FR-11: Map well-known types:
  - `google.protobuf.Timestamp` -> `TIMESTAMPTZ`
  - `google.protobuf.Duration` -> `INTERVAL`
  - `google.protobuf.Struct` / `google.protobuf.Value` -> `JSONB`
- [x] FR-12: Map enums to `TEXT` with a `CHECK` constraint listing the valid proto enum value names
- [x] FR-13: Map `repeated <scalar>` to PostgreSQL array types (e.g. `TEXT[]`, `INTEGER[]`)
- [x] FR-14: Map `repeated <message>`, nested messages, and `map<K,V>` to `JSONB`
- [x] FR-15: Map `oneof` to nullable columns, one per variant

### Non-Functional

- [x] NFR-1: UUIDs are generated at the application layer using `gofrs/uuid/v5` — no database-level UUID generation
- [x] NFR-2: Generated sqlc config targets PostgreSQL with the pgx driver

### Deferred

- [ ] DFR-1: Schema migration support (diffing existing schema, generating ALTER statements) — initial version generates fresh DDL only
- [ ] DFR-2: Custom column type overrides via proto options
- [ ] DFR-3: Index generation from proto annotations

## Constraints

- Must work as a standard protoc plugin (`protoc-gen-clarity`)
- Output directory structure follows `internal/<provider>/<domain>/<version>/` convention
- Single `schema.sql` per package, separate query file per entity

## Out of Scope

- Database migration tooling (up/down migrations)
- Support for databases other than PostgreSQL
- Foreign key relationships between entities
- Custom query generation beyond CRUD
