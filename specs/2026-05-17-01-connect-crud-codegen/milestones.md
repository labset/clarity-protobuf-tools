# Milestones: Connect CRUD Codegen

## M1: Generator scaffolding and handler template

- **Status:** pending
- **Description:** Register `connect-crud` mode, add `generator_connect_crud.go`, produce `handler_<model>.go` with the `<Model>Deps` struct, constructor (creates SQLC store from pool), and Connect service registration. No RPC implementations yet.
- **Acceptance Criteria:**
  - [ ] `connect-crud` mode is registered and selectable via plugin options
  - [ ] `generator_connect_crud.go` follows existing generator conventions
  - [ ] `handler_<model>.go` is generated under `<provider>/<domain>/<version>/api/`
  - [ ] Handler struct holds the SQLC store, created once in the constructor from `*pgxpool.Pool` in deps
  - [ ] Connect service handler registration wires up to the handler struct

## M2: Mapper generation

- **Status:** pending
- **Description:** Generate `mapper_<model>.go` per entity with proto-to-SQLC and SQLC-to-proto conversion functions for all field types.
- **Acceptance Criteria:**
  - [ ] `mapper_<model>.go` is generated under `<provider>/<domain>/<version>/api/`
  - [ ] UUID string ↔ `uuid.UUID` conversion
  - [ ] `google.protobuf.Timestamp` ↔ `time.Time` conversion
  - [ ] Standard field type mappings (string, int32, bool, etc.)
  - [ ] Handles nullable/optional fields

## M3: Create and Get operations

- **Status:** pending
- **Description:** Generate `rpc_create_<model>.go` and `rpc_get_<model>.go` with full implementations including store calls, mapping, and error handling.
- **Acceptance Criteria:**
  - [ ] `rpc_create_<model>.go` maps request to SQLC params, calls store insert, maps result to proto response
  - [ ] `rpc_get_<model>.go` calls store get-by-id, returns `CodeNotFound` if not found
  - [ ] Soft-delete filtering is at the store layer, not in the handler
  - [ ] Consistent Connect error codes for failure cases

## M4: List operation with cursor-based pagination

- **Status:** pending
- **Description:** Generate `rpc_list_<model>.go` with cursor-based pagination using `page_size` and `page_token`.
- **Acceptance Criteria:**
  - [ ] `rpc_list_<model>.go` is generated with cursor encode/decode logic
  - [ ] Respects `page_size` from request
  - [ ] Decodes `page_token` for cursor position, encodes `next_page_token` in response
  - [ ] Calls SQLC store list method with appropriate parameters
  - [ ] Returns `CodeInvalidArgument` for malformed page tokens

## M5: Update and Delete operations

- **Status:** pending
- **Description:** Generate `rpc_update_<model>.go` with field mask handling and `rpc_delete_<model>.go` for soft delete.
- **Acceptance Criteria:**
  - [ ] `rpc_update_<model>.go` respects `update_mask` for partial updates
  - [ ] `rpc_update_<model>.go` returns `CodeNotFound` if entity does not exist
  - [ ] `rpc_delete_<model>.go` calls store soft-delete (sets `deleted_at`)
  - [ ] `rpc_delete_<model>.go` returns `CodeNotFound` if entity does not exist

## M6: Golden file tests and validation

- **Status:** pending
- **Description:** Add golden file tests for all generated files, ensure generated code compiles and passes `go vet`.
- **Acceptance Criteria:**
  - [ ] Golden files exist under `internal/codegen/testdata/golden/connect-crud/`
  - [ ] Tests cover handler, mapper, and all RPC operation files
  - [ ] Generated code compiles successfully
  - [ ] Generated code passes `go vet`
