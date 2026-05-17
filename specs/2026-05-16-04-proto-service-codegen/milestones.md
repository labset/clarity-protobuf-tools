# Milestones: Proto Service Codegen

## M1: Extend proto options with operations

- **Status:** pending
- **Description:** Add the `Operation` enum and `operations` field to `ClarityMessageOptions` in `options.proto`, regenerate Go code
- **Acceptance Criteria:**
  - [ ] `Operation` enum defined with `UNSPECIFIED`, `CREATE`, `GET`, `LIST`, `UPDATE`, `DELETE`
  - [ ] `ClarityMessageOptions` has a `repeated Operation operations` field
  - [ ] Generated Go code compiles and existing tests pass

## M2: Service codegen generator scaffold

- **Status:** pending
- **Description:** Add `generator_service.go` with the `service` mode wired into the plugin, producing output directory structure and empty files
- **Acceptance Criteria:**
  - [ ] `service` mode accepted by `protoc-gen-clarity`
  - [ ] Output directory follows `<provider>/<domain>/<version>/` structure
  - [ ] Generates `service_<model>.proto` and `rpc_<op>_<model>.proto` file stubs per entity with operations

## M3: Generate RPC request/response protos

- **Status:** pending
- **Description:** Implement templates for each operation's request/response message file (`rpc_<op>_<model>.proto`)
- **Acceptance Criteria:**
  - [ ] `rpc_create_<model>.proto` with request (entity fields minus id/timestamps) and response (wraps entity)
  - [ ] `rpc_get_<model>.proto` with request (`string id`) and response (wraps entity)
  - [ ] `rpc_list_<model>.proto` with request (page_size, page_token) and response (repeated entity, next_page_token)
  - [ ] `rpc_update_<model>.proto` with request (id + mutable fields) and response (wraps entity)
  - [ ] `rpc_delete_<model>.proto` with request (`string id`) and response (empty/ack)
  - [ ] All generated files include correct syntax, package, go_package, and imports

## M4: Generate service proto

- **Status:** pending
- **Description:** Implement the `service_<model>.proto` template that defines the service with RPCs referencing the generated request/response messages
- **Acceptance Criteria:**
  - [ ] Service proto imports all relevant `rpc_<op>_<model>.proto` files
  - [ ] Service definition lists only opted-in operations as RPCs
  - [ ] Generated proto is valid (`buf lint` / `protoc` parseable)

## M5: Golden file tests and validation

- **Status:** pending
- **Description:** Add golden file tests for the service generator covering various operation combinations
- **Acceptance Criteria:**
  - [ ] Golden files under `internal/codegen/testdata/golden/service/`
  - [ ] Test cases for: single operation, all operations, multiple entities
  - [ ] Generated protos pass `buf lint` validation in CI
