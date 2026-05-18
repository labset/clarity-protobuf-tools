# Milestones: River Outbox Codegen

## M1: Outbox event codegen

- **Status:** done
- **Description:** Generate River `JobArgs` event structs for each mutating CRUD operation. This milestone delivers the `outbox/event_<op>_<model>.go` files with correct `Kind()` methods and payload fields.
- **Acceptance Criteria:**
  - [x] New `connect-crud-outbox` mode registered in the codegen plugin
  - [x] Event templates created under `internal/codegen/templates/connect-crud-outbox/`
  - [x] Create event args generated with `EntityID` and `OccurredAt`
  - [x] Update event args generated with `EntityID`, `FieldMask`, and `OccurredAt`
  - [x] Delete event args generated with `EntityID` and `OccurredAt`
  - [x] `Kind()` returns `<op>_<model_snake>` (e.g. `create_book`)
  - [x] Golden file tests pass for generated event files

## M2: Transactional CRUD handler codegen

- **Status:** done
- **Description:** Generate Connect RPC handlers that wrap Create/Update/Delete operations in a Postgres transaction and atomically enqueue the corresponding River outbox event. Get/List handlers remain unchanged.
- **Acceptance Criteria:**
  - [x] Handler deps include `*river.Client[pgx.Tx]` alongside `*pgxpool.Pool`
  - [x] Create/Update/Delete RPCs begin a transaction, execute the SQLC query, call `river.InsertTx`, and commit
  - [x] Get and List RPCs remain non-transactional (same as `connect-crud`)
  - [x] Mapper generation reused from existing `connect-crud` logic
  - [x] Golden file tests pass for all generated handler and RPC files

## M3: Composition and integration

- **Status:** done
- **Description:** Wire the `connect-crud-outbox` generator to compose the `sqlc` and `atlas` generators, ensuring the full codegen pipeline produces all expected outputs (migrations, queries, store, handlers, outbox events).
- **Acceptance Criteria:**
  - [x] `connect-crud-outbox` mode composes `atlasSqlcGenerator` like `connect-crud` does
  - [x] End-to-end golden file test validates the complete output tree
  - [x] Build passes (`mise run build`)
