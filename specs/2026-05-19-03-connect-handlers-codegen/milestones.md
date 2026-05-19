# Milestones: Connect Handlers Codegen

## M1: Generator and templates

- **Status:** pending
- **Description:** Implement the `connect-handlers` generator, templates for `handler_<model>.go` and `rpc_<name>.go`, and the "skip if exists" logic.
- **Acceptance Criteria:**
  - [ ] `generator_connect_handlers.go` registered as a codegen mode
  - [ ] Templates under `internal/codegen/templates/connect-handlers/`
  - [ ] `handler_<model>.go` template generates deps struct, constructor, and Connect service handler registration
  - [ ] `rpc_<name>.go` template generates stub returning `connect.CodeUnimplemented`
  - [ ] Files are only emitted when they do not already exist on disk

## M2: Golden file tests

- **Status:** pending
- **Description:** Add golden file tests using a test service proto with multiple RPCs, verifying the generated output.
- **Acceptance Criteria:**
  - [ ] Golden files under `internal/codegen/testdata/golden/connect-handlers/`
  - [ ] Test covers a service with multiple RPC methods
  - [ ] `mise run go:test` passes

## M3: E2E test

- **Status:** pending
- **Description:** Add an end-to-end test with a buf plugin config that runs the full generation pipeline and verifies the output compiles.
- **Acceptance Criteria:**
  - [ ] E2E test directory under `test/connect-handlers/`
  - [ ] buf plugin config invokes `connect-handlers` mode
  - [ ] Generated output compiles and passes `go vet`
  - [ ] `mise run build` passes
