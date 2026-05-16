# Milestones: data-model-codegen

## M1: File convention enforcement

- **Status:** pending
- **Description:** Codegen plugin only processes `models.proto` files, and the lint plugin rejects `ROLE_ENTITY` annotations on messages defined in any other file.
- **Acceptance Criteria:**
  - [ ] Codegen skips `.proto` files not named `models.proto`
  - [ ] Lint rule rejects `ROLE_ENTITY` on messages in files not named `models.proto`
  - [ ] Existing behaviour unchanged for `models.proto` files
  - [ ] Unit tests cover both skip and reject cases

## M2: Soft delete and timestamp immutability

- **Status:** pending
- **Description:** Inline a nullable `deleted_at TIMESTAMPTZ` column on all entity tables, generate a `SoftDelete<Entity>` query, filter deleted rows from Get and List, and fix the Update query to exclude `created_at`/`deleted_at` and use `NOW()` for `updated_at`.
- **Acceptance Criteria:**
  - [ ] `deleted_at TIMESTAMPTZ` inlined as a nullable column in schema DDL
  - [ ] `SoftDelete<Entity>` query sets `deleted_at = NOW()` by id
  - [ ] `Get<Entity>` includes `WHERE deleted_at IS NULL`
  - [ ] `List<Entity>s` includes `WHERE deleted_at IS NULL`
  - [ ] `Delete<Entity>` (hard delete) remains unchanged
  - [ ] `Update<Entity>` SET clause excludes `created_at` and `deleted_at`
  - [ ] `Update<Entity>` uses `NOW()` for `updated_at`
  - [ ] Golden files updated and tests pass

## M3: Query and schema improvements

- **Status:** pending
- **Description:** Use `RETURNING *` on Insert and Update queries, add `IF NOT EXISTS` to `CREATE TABLE` statements, and derive sqlc.yaml `package` and `out` from the proto domain.
- **Acceptance Criteria:**
  - [ ] `Create<Entity>` uses `:one` with `RETURNING *`
  - [ ] `Update<Entity>` uses `:one` with `RETURNING *`
  - [ ] `CREATE TABLE` statements include `IF NOT EXISTS`
  - [ ] sqlc.yaml `package` and `out` derived from proto domain
  - [ ] Golden files updated and tests pass

## M4: Cross-file aggregation and wrapper types

- **Status:** pending
- **Description:** Aggregate entity messages across multiple `.proto` files in the same package into a single `schema.sql`, and map Google wrapper types to nullable scalar columns instead of JSONB.
- **Acceptance Criteria:**
  - [ ] Multiple `models.proto` files in the same package produce a single `schema.sql`
  - [ ] `google.protobuf.StringValue` maps to nullable `TEXT`
  - [ ] `google.protobuf.Int32Value` maps to nullable `INTEGER`
  - [ ] `google.protobuf.Int64Value` maps to nullable `BIGINT`
  - [ ] `google.protobuf.BoolValue` maps to nullable `BOOLEAN`
  - [ ] `google.protobuf.FloatValue` maps to nullable `REAL`
  - [ ] `google.protobuf.DoubleValue` maps to nullable `DOUBLE PRECISION`
  - [ ] `google.protobuf.UInt32Value` maps to nullable `INTEGER`
  - [ ] `google.protobuf.UInt64Value` maps to nullable `BIGINT`
  - [ ] `google.protobuf.BytesValue` maps to nullable `BYTEA`
  - [ ] Golden files and unit tests cover all wrapper type mappings
