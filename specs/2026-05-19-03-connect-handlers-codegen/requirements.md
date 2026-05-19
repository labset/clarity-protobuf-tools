# Requirements: Connect Handlers Codegen

> Generate Connect-RPC handler scaffolding with stub implementations for any proto service definition.

## Context

The existing `connect-crud` and `connect-crud-outbox` codegen modes generate fully wired handler implementations for CRUD operations backed by SQLC stores. However, not all services follow CRUD patterns — custom RPCs need handler scaffolding too. This mode provides a one-time scaffold generator that reads proto service definitions and emits handler stubs returning `Unimplemented`, giving developers a starting point to fill in business logic.

## Requirements

### Functional

- [x] FR-1: Add a new `connect-handlers` codegen mode to `protoc-gen-clarity`
- [x] FR-2: Process any proto service definition (not limited to `models.proto`)
- [x] FR-3: Generate output under `<output_dir>/internal/<provider>/<domain>/<version>/api/`
- [x] FR-4: Generate `handler_<model>.go` per service containing: a deps struct, a constructor, and Connect service handler registration
- [x] FR-5: Generate `rpc_<name>.go` per RPC method containing a stub implementation that returns `connect.CodeUnimplemented`
- [x] FR-6: Only emit files that do not already exist — never overwrite existing handler or RPC files

### Non-Functional

- [x] NFR-1: Follow existing codegen conventions — `generator_` file prefix, embedded templates, golden file tests
- [x] NFR-2: Templates live under `internal/codegen/templates/connect-handlers/`
- [x] NFR-3: Golden test files scoped under `internal/codegen/testdata/golden/connect-handlers/`

### Deferred

- [ ] DFR-1: Generating mapper files for request/response types
- [ ] DFR-2: Generating test scaffolding alongside handler stubs
- [ ] DFR-3: Option to regenerate/update stubs when new RPCs are added to an existing service

## Constraints

- Must remain a standard protoc/buf plugin
- Standalone mode — does not compose other generators
- Generated files are scaffolding — intended to be edited by developers, not regenerated

## Out of Scope

- SQLC store wiring or database integration
- Outbox/event generation
- Middleware or interceptor integration
- Full handler implementations
