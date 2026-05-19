# Milestones: Connect Handlers Codegen

## M1: Generator and templates

- **Status:** done
- **Description:** Implement the `connect-handlers` generator, templates for `handler_<model>.go` and `rpc_<name>.go`, and the "skip if exists" logic.
- **Acceptance Criteria:**
  - [x] `generator_connect_handlers.go` registered as a codegen mode
  - [x] Templates under `internal/codegen/templates/connect-handlers/`
  - [x] `handler_<model>.go` template generates deps struct, constructor, and Connect service handler registration
  - [x] `rpc_<name>.go` template generates stub returning `connect.CodeUnimplemented`
  - [x] Files are only emitted when they do not already exist on disk

## M2: Golden file tests

- **Status:** done
- **Description:** Add golden file tests using a test service proto with multiple RPCs, verifying the generated output.
- **Acceptance Criteria:**
  - [x] Golden files under `internal/codegen/testdata/golden/connect-handlers/`
  - [x] Test covers a service with multiple RPC methods
  - [x] `mise run go:test` passes

## M3: E2E test

- **Status:** done
- **Description:** Add an end-to-end test with a buf plugin config that runs the full generation pipeline and verifies the output compiles.
- **Acceptance Criteria:**
  - [x] E2E test directory under `test/connect-handlers/`
  - [x] buf plugin config invokes `connect-handlers` mode
  - [x] Generated output compiles and passes `go vet`
  - [x] `mise run build` passes
