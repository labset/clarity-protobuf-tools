# Requirements: data-model-codegen

> Harden the sqlc codegen plugin with file conventions, soft delete support, timestamp immutability, and query/schema improvements.

## Context

The initial sqlc codegen spec delivered working DDL and CRUD query generation from proto entity annotations. This follow-up addresses correctness issues, developer experience gaps, and conventions discovered during review. The `deleted_at` column is managed purely at the SQL layer — it is not part of the `Entity` proto message.

## Requirements

### Functional

- [x] FR-1: Codegen plugin only processes files named `models.proto` — all other `.proto` files are skipped
- [x] FR-2: Lint plugin rejects `ROLE_ENTITY` annotation on messages not in a file named `models.proto`
- [x] FR-3: Inline `deleted_at TIMESTAMPTZ` as a nullable column on all entity tables (not in the proto `Entity` message)
- [x] FR-4: Generate a `SoftDelete<Entity>` query that sets `deleted_at = NOW()` by id
- [x] FR-5: `Get<Entity>` query filters `WHERE deleted_at IS NULL`
- [x] FR-6: `List<Entity>s` query filters `WHERE deleted_at IS NULL`
- [x] FR-7: `Delete<Entity>` (hard delete) query remains available
- [x] FR-8: `Update<Entity>` excludes `created_at` and `deleted_at` from the SET clause
- [x] FR-9: `Update<Entity>` uses `NOW()` for `updated_at` instead of a caller-provided parameter
- [x] FR-10: `Create<Entity>` and `Update<Entity>` use `RETURNING *` (sqlc `:one` instead of `:exec`)
- [x] FR-11: Add `IF NOT EXISTS` to `CREATE TABLE` statements
- [x] FR-12: sqlc.yaml `package` and `out` use `"db"` convention
- [x] FR-13: Aggregate entity messages across multiple `.proto` files in the same package into a single `schema.sql`
- [x] FR-14: Map Google wrapper types (`StringValue`, `Int32Value`, `BoolValue`, etc.) to nullable scalar columns instead of JSONB
- [x] FR-15: Lint plugin validates `ROLE_ENTITY` messages are in a `<provider>.<domain>.<version>` package
- [x] FR-16: Extract shared option extraction logic into `internal/clarity` package
- [x] FR-17: Exclude `deleted_at` from INSERT columns — always defaults to `NULL` on creation
- [x] FR-18: `Update<Entity>` and `SoftDelete<Entity>` guard with `WHERE deleted_at IS NULL` — prevent operating on soft-deleted rows
- [x] FR-19: Use sqlc named args (`@column_name`) instead of positional parameters (`$1`, `$2`)

### Non-Functional

- [x] NFR-1: All generated SQL must pass `sqlc compile` validation
- [x] NFR-2: Golden file tests updated for all changed output
- [x] NFR-3: Full test coverage for `CLARITY_ENTITY_FIELD` lint rule (pass, missing field, wrong number, wrong type)

### Deferred

- [ ] DFR-1: Custom column type overrides via proto options
- [ ] DFR-2: Index generation from proto annotations
- [ ] DFR-3: Uniqueness constraints via proto annotations
- [ ] DFR-4: Cursor-based pagination for List queries
- [ ] DFR-5: Smarter pluralization for List query names

## Constraints

- `deleted_at` is codegen-only — the `Entity` proto message is not modified
- No new proto option annotations in this spec
- Must remain a standard protoc plugin

## Out of Scope

- Schema migration / ALTER statement generation
- Foreign key relationships
- Databases other than PostgreSQL
- Custom query generation beyond CRUD + soft delete
