# Milestones: River Outbox Codegen

## M1: Outbox event codegen

- **Status:** pending
- **Description:** Generate River `JobArgs` event structs for each mutating CRUD operation. This milestone delivers the `outbox/event_<op>_<model>.go` files with correct `Kind()` methods and payload fields.
- **Acceptance Criteria:**
  - [ ] New `connect-crud-outbox` mode registered in the codegen plugin
  - [ ] Event templates created under `internal/codegen/templates/connect-crud-outbox/`
  - [ ] Create event args generated with `EntityID` and `OccurredAt`
  - [ ] Update event args generated with `EntityID`, `FieldMask`, and `OccurredAt`
  - [ ] Delete event args generated with `EntityID` and `OccurredAt`
  - [ ] `Kind()` returns `<op>_<model_snake>` (e.g. `create_book`)
  - [ ] Golden file tests pass for generated event files

## M2: Transactional CRUD handler codegen

- **Status:** pending
- **Description:** Generate Connect RPC handlers that wrap Create/Update/Delete operations in a Postgres transaction and atomically enqueue the corresponding River outbox event. Get/List handlers remain unchanged.
- **Acceptance Criteria:**
  - [ ] Handler deps include `*river.Client[pgx.Tx]` alongside `*pgxpool.Pool`
  - [ ] Create/Update/Delete RPCs begin a transaction, execute the SQLC query, call `river.InsertTx`, and commit
  - [ ] Get and List RPCs remain non-transactional (same as `connect-crud`)
  - [ ] Mapper generation reused from existing `connect-crud` logic
  - [ ] Golden file tests pass for all generated handler and RPC files

## M3: Composition and integration

- **Status:** pending
- **Description:** Wire the `connect-crud-outbox` generator to compose the `sqlc` and `atlas` generators, ensuring the full codegen pipeline produces all expected outputs (migrations, queries, store, handlers, outbox events).
- **Acceptance Criteria:**
  - [ ] `connect-crud-outbox` mode composes `atlasSqlcGenerator` like `connect-crud` does
  - [ ] End-to-end golden file test validates the complete output tree
  - [ ] Build passes (`mise run build`)
