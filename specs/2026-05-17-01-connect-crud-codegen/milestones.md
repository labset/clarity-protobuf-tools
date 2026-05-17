# Milestones: Connect CRUD Codegen

## M1: Generator scaffolding and handler template

- **Status:** done
- **Description:** Register `connect-crud` mode, add `generator_connect_crud.go`, produce `handler_<model>.go` with the `<Model>Deps` struct, constructor (creates SQLC store from pool), and Connect service registration. No RPC implementations yet.
- **Acceptance Criteria:**
  - [x] `connect-crud` mode is registered and selectable via plugin options
  - [x] `generator_connect_crud.go` follows existing generator conventions
  - [x] `handler_<model>.go` is generated under `<provider>/<domain>/<version>/api/`
  - [x] Handler struct holds the SQLC store, created once in the constructor from `*pgxpool.Pool` in deps
  - [x] Connect service handler registration wires up to the handler struct

## M2: Mapper generation

- **Status:** done
- **Description:** Generate `mapper_<model>.go` per entity with proto-to-SQLC and SQLC-to-proto conversion functions for all field types.
- **Acceptance Criteria:**
  - [x] `mapper_<model>.go` is generated under `<provider>/<domain>/<version>/api/`
  - [x] UUID string ↔ `uuid.UUID` conversion
  - [x] `google.protobuf.Timestamp` ↔ `time.Time` conversion
  - [x] Standard field type mappings (string, int32, bool, etc.)

## M3: Create and Get operations

- **Status:** done
- **Description:** Generate `rpc_create_<model>.go` and `rpc_get_<model>.go` with full implementations including store calls, mapping, and error handling.
- **Acceptance Criteria:**
  - [x] `rpc_create_<model>.go` maps request to SQLC params, calls store insert, maps result to proto response
  - [x] `rpc_get_<model>.go` calls store get-by-id, returns `CodeNotFound` if not found
  - [x] Soft-delete filtering is at the store layer, not in the handler
  - [x] Consistent Connect error codes for failure cases

## M4: List operation with cursor-based pagination

- **Status:** done
- **Description:** Generate `rpc_list_<model>.go` with cursor-based pagination using `page_size` and `page_token`.
- **Acceptance Criteria:**
  - [x] `rpc_list_<model>.go` is generated with cursor encode/decode logic
  - [x] Respects `page_size` from request
  - [x] Decodes `page_token` for cursor position, encodes `next_page_token` in response
  - [x] Calls SQLC store list method with appropriate parameters
  - [x] Returns `CodeInvalidArgument` for malformed page tokens

## M5: Update and Delete operations

- **Status:** done
- **Description:** Generate `rpc_update_<model>.go` with field mask handling and `rpc_delete_<model>.go` for soft delete.
- **Acceptance Criteria:**
  - [x] `rpc_update_<model>.go` respects `update_mask` for partial updates
  - [x] `rpc_update_<model>.go` returns `CodeNotFound` if entity does not exist
  - [x] `rpc_delete_<model>.go` calls store soft-delete (sets `deleted_at`)
  - [x] `rpc_delete_<model>.go` returns `CodeNotFound` if entity does not exist

## M6: Golden file tests and validation

- **Status:** done
- **Description:** Add golden file tests for all generated files, ensure generated code compiles and passes `go vet`.
- **Acceptance Criteria:**
  - [x] Golden files exist under `internal/codegen/testdata/golden/connect-crud/`
  - [x] Tests cover handler, mapper, and all RPC operation files
  - [x] Generated code compiles successfully
  - [x] Generated code passes `go vet`
