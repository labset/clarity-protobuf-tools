# Milestones: data-model-codegen

## M1: File convention enforcement

- **Status:** done
- **Description:** Codegen plugin only processes `models.proto` files, and the lint plugin rejects `ROLE_ENTITY` annotations on messages defined in any other file.
- **Acceptance Criteria:**
  - [x] Codegen skips `.proto` files not named `models.proto`
  - [x] Lint rule rejects `ROLE_ENTITY` on messages in files not named `models.proto`
  - [x] Existing behaviour unchanged for `models.proto` files
  - [x] Unit tests cover both skip and reject cases

## M2: Soft delete and timestamp immutability

- **Status:** done
- **Description:** Inline a nullable `deleted_at TIMESTAMPTZ` column on all entity tables, generate a `SoftDelete<Entity>` query, filter deleted rows from Get and List, and fix the Update query to exclude `created_at`/`deleted_at` and use `NOW()` for `updated_at`.
- **Acceptance Criteria:**
  - [x] `deleted_at TIMESTAMPTZ` inlined as a nullable column in schema DDL
  - [x] `SoftDelete<Entity>` query sets `deleted_at = NOW()` by id
  - [x] `Get<Entity>` includes `WHERE deleted_at IS NULL`
  - [x] `List<Entity>s` includes `WHERE deleted_at IS NULL`
  - [x] `Delete<Entity>` (hard delete) remains unchanged
  - [x] `Update<Entity>` SET clause excludes `created_at` and `deleted_at`
  - [x] `Update<Entity>` uses `NOW()` for `updated_at`
  - [x] Golden files updated and tests pass

## M3: Query and schema improvements

- **Status:** done
- **Description:** Use `RETURNING *` on Insert and Update queries, add `IF NOT EXISTS` to `CREATE TABLE` statements, and derive sqlc.yaml `package` and `out` from the proto domain.
- **Acceptance Criteria:**
  - [x] `Create<Entity>` uses `:one` with `RETURNING *`
  - [x] `Update<Entity>` uses `:one` with `RETURNING *`
  - [x] `CREATE TABLE` statements include `IF NOT EXISTS`
  - [x] sqlc.yaml `package` and `out` derived from proto domain
  - [x] Golden files updated and tests pass

## M4: Cross-file aggregation and wrapper types

- **Status:** done
- **Description:** Generator groups entity messages by package (supporting multiple packages in a single plugin run), and maps Google wrapper types to nullable scalar columns instead of JSONB. Note: cross-file aggregation within the same package is moot given M1's `models.proto` convention (one file per package), but the grouping logic is in place.
- **Acceptance Criteria:**
  - [x] Generator groups files by package, producing one `schema.sql` per package
  - [x] `google.protobuf.StringValue` maps to nullable `TEXT`
  - [x] `google.protobuf.Int32Value` maps to nullable `INTEGER`
  - [x] `google.protobuf.Int64Value` maps to nullable `BIGINT`
  - [x] `google.protobuf.BoolValue` maps to nullable `BOOLEAN`
  - [x] `google.protobuf.FloatValue` maps to nullable `REAL`
  - [x] `google.protobuf.DoubleValue` maps to nullable `DOUBLE PRECISION`
  - [x] `google.protobuf.UInt32Value` maps to nullable `INTEGER`
  - [x] `google.protobuf.UInt64Value` maps to nullable `BIGINT`
  - [x] `google.protobuf.BytesValue` maps to nullable `BYTEA`
  - [x] Golden files and unit tests cover all wrapper type mappings
