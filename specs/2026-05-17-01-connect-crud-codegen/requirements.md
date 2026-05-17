# Requirements: Connect CRUD Codegen

> Generate Go Connect-RPC handler implementations for CRUD operations, backed by SQLC-generated stores, from proto entity annotations.

## Context

The `protoc-gen-clarity` plugin already generates proto service definitions (`service` mode), SQLC queries/models (`sqlc` mode), and Atlas migrations (`atlas` mode). The final piece is generating the Go handler layer that wires Connect-RPC services to the SQLC stores. This eliminates boilerplate handler code and ensures consistency between the API and data layers.

## Requirements

### Functional

- [ ] FR-1: Add a new `connect-crud` codegen mode to `protoc-gen-clarity`
- [ ] FR-2: Generate output under `<output_dir>/<provider>/<domain>/<version>/api/`
- [ ] FR-3: Generate `handler_<model>.go` per entity containing: a `<Model>Deps` struct (with `*pgxpool.Pool`), a constructor that creates the SQLC store, and Connect service handler registration
- [ ] FR-4: Generate `rpc_<op>_<model>.go` per opted-in operation containing the method implementation
- [ ] FR-5: Generate `mapper_<model>.go` per entity with proto-to-SQLC and SQLC-to-proto conversion functions (UUID string ↔ `uuid.UUID`, `google.protobuf.Timestamp` ↔ `time.Time`, etc.)
- [ ] FR-6: `Create` handler calls the SQLC store's insert method, maps request to store params, maps result back to proto response
- [ ] FR-7: `Get` handler calls the SQLC store's get-by-id method (with soft-delete filter at store layer), returns NotFound if no result
- [ ] FR-8: `List` handler implements cursor-based pagination using `page_size` and `page_token`, calling the SQLC store's list method
- [ ] FR-9: `Update` handler calls the SQLC store's update method respecting `update_mask`, returns NotFound if no result
- [ ] FR-10: `Delete` handler calls the SQLC store's soft-delete method (sets `deleted_at`), returns NotFound if no result
- [ ] FR-11: Consistent Connect error handling: `CodeNotFound` for missing entities, `CodeInvalidArgument` for validation failures, `CodeAlreadyExists` for conflicts
- [ ] FR-12: Import the SQLC store package using the path derived from the `sqlc` codegen mode output
- [ ] FR-13: Generated handlers implement the Connect service interface generated from the proto service definitions

### Non-Functional

- [ ] NFR-1: Follow existing codegen conventions — `generator_` file prefix, embedded templates, golden file tests
- [ ] NFR-2: Only process entities in files named `models.proto` (consistent with existing modes)
- [ ] NFR-3: Generated Go files must compile and pass `go vet`

### Deferred

- [ ] DFR-1: Middleware/interceptor integration (auth, logging, tracing)
- [ ] DFR-2: Custom RPC operations beyond CRUD
- [ ] DFR-3: Filtering/sorting options on List operations
- [ ] DFR-4: Bulk operations (batch create, batch delete)
- [ ] DFR-5: Hard delete / purge RPC
- [ ] DFR-6: Nullable/optional field handling in mapper conversions

## Constraints

- Must remain a standard protoc/buf plugin
- Delete always means soft delete at the handler layer
- Operations are opt-in — entities without `operations` produce no output
- SQLC store import path is derived from the `sqlc` mode output, not configured separately

## Out of Scope

- Proto service definition generation (handled by `service` mode)
- SQL query generation (handled by `sqlc` mode)
- Migration generation (handled by `atlas` mode)
- gRPC transport (Connect-only)
- Authentication/authorization logic
- Runtime pagination cursor encoding strategy (implementation detail)
