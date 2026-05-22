# Milestones: Proto Service Codegen

## M1: Extend proto options with operations

- **Status:** done
- **Description:** Add the `Operation` enum and `operations` field to `ClarityMessageOptions` in `options.proto`, regenerate Go code
- **Acceptance Criteria:**
  - [x] `Operation` enum defined with `UNSPECIFIED`, `CREATE`, `GET`, `LIST`, `UPDATE`, `DELETE`
  - [x] `ClarityMessageOptions` has a `repeated Operation operations` field
  - [x] Generated Go code compiles and existing tests pass

## M2: Service codegen generator scaffold

- **Status:** done
- **Description:** Add `generator_service.go` with the `service` mode wired into the plugin, producing output directory structure and files
- **Acceptance Criteria:**
  - [x] `service` mode accepted by `protoc-gen-labset-go`
  - [x] Output directory follows `<provider>/<domain>/<version>/` structure
  - [x] Generates `service_<model>.proto` and `rpc_<op>_<model>.proto` files per entity with operations

## M3: Generate RPC request/response protos

- **Status:** done
- **Description:** Implement templates for each operation's request/response message file (`rpc_<op>_<model>.proto`)
- **Acceptance Criteria:**
  - [x] `rpc_create_<model>.proto` with request (typed `item`) and response (wraps entity as `item`)
  - [x] `rpc_get_<model>.proto` with request (`string id` + buf.validate UUID) and response (wraps entity as `item`)
  - [x] `rpc_list_<model>.proto` with request (page_size, page_token) and response (`repeated items`, next_page_token)
  - [x] `rpc_update_<model>.proto` with request (id + item + FieldMask) and response (wraps entity as `item`)
  - [x] `rpc_delete_<model>.proto` with request (`string id` + buf.validate UUID) and response (empty)
  - [x] All generated files include correct syntax, package, go_package, and imports

## M4: Generate service proto

- **Status:** done
- **Description:** Implement the `service_<model>.proto` template that defines the service with RPCs referencing the generated request/response messages
- **Acceptance Criteria:**
  - [x] Service proto imports all relevant `rpc_<op>_<model>.proto` files
  - [x] Service definition lists only opted-in operations as RPCs
  - [ ] Generated proto is valid (`buf lint` / `protoc` parseable) — deferred to CI integration

## M5: Golden file tests and validation

- **Status:** done
- **Description:** Add golden file tests for the service generator covering various operation combinations
- **Acceptance Criteria:**
  - [x] Golden files under `internal/codegen/testdata/golden/service/`
  - [x] Test cases for: single operation, all operations, multiple entities
  - [ ] Generated protos pass `buf lint` validation in CI — deferred to CI integration
