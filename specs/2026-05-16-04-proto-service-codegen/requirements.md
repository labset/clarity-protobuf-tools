# Requirements: Proto Service Codegen

> Generate `.proto` service definitions (RPCs + request/response messages) from entity annotations with opt-in operation selection.

## Context

The `protoc-gen-labset-go` plugin currently generates SQL schemas and queries from proto entity definitions. To complete the developer workflow, we need to generate proto service definitions that expose CRUD operations for annotated entities. This keeps the API layer consistent with the data model and avoids hand-writing boilerplate service protos.

Operations are opt-in per entity via the existing `ClarityMessageOptions`, and the generated protos follow the same package structure as the source models.

## Requirements

### Functional

- [x] FR-1: Add `Operation` enum to `clarity.plugin.v1` with values: `OPERATION_UNSPECIFIED`, `OPERATION_CREATE`, `OPERATION_GET`, `OPERATION_LIST`, `OPERATION_UPDATE`, `OPERATION_DELETE`
- [x] FR-2: Add repeated `operations` field to `ClarityMessageOptions`
- [x] FR-3: Add a new `service` codegen mode to `protoc-gen-labset-go`
- [x] FR-4: Generate output under `<output_dir>/<provider>/<domain>/<version>/` respecting proto package structure
- [x] FR-5: Generate `service_<model>.proto` per entity containing a `service <Model>Service` with RPCs for each opted-in operation
- [x] FR-6: Generate `rpc_<op>_<model>.proto` per operation per entity containing request and response messages
- [x] FR-7: `OPERATION_CREATE` generates `rpc Create<Model>(Create<Model>Request) returns (Create<Model>Response)` — request contains typed `item` field; response wraps the entity as `item`
- [x] FR-8: `OPERATION_GET` generates `rpc Get<Model>(Get<Model>Request) returns (Get<Model>Response)` — request contains `string id` with `buf.validate` UUID constraint; response wraps the entity as `item`
- [x] FR-9: `OPERATION_LIST` generates `rpc List<Model>s(List<Model>sRequest) returns (List<Model>sResponse)` — request contains `int32 page_size` and `string page_token`; response contains `repeated <Model> items` and `string next_page_token`
- [x] FR-10: `OPERATION_UPDATE` generates `rpc Update<Model>(Update<Model>Request) returns (Update<Model>Response)` — request contains `string id` (validated), typed `item` field, and `google.protobuf.FieldMask update_mask`; response wraps the entity as `item`
- [x] FR-11: `OPERATION_DELETE` generates `rpc Delete<Model>(Delete<Model>Request) returns (Delete<Model>Response)` — semantics are always soft delete; request contains `string id` with `buf.validate` UUID constraint; response is empty
- [x] FR-12: Generated protos import the entity message from the same proto package using full path (`<provider>/<domain>/<version>/models.proto`)
- [x] FR-13: Generated protos include correct `syntax`, `package`, and `option go_package` declarations matching the source
- [x] FR-14: `id` fields on get, update, and delete requests use `buf.validate` UUID validation
- [x] FR-15: Update request includes `google.protobuf.FieldMask update_mask` for partial updates

### Non-Functional

- [x] NFR-1: Follow existing codegen conventions — `generator_` file prefix, embedded templates, golden file tests
- [x] NFR-2: Only process entities in files named `models.proto` (consistent with existing modes)
- [ ] NFR-3: Generated `.proto` files must be valid and parseable by `buf lint` / `protoc`

### Deferred

- [ ] DFR-1: Go handler/stub generation from the generated service protos
- [ ] DFR-2: Hard delete / purge maintenance RPC
- [ ] DFR-3: Custom RPC operations beyond CRUD
- [ ] DFR-4: Filtering/sorting options on List operations

## Constraints

- Must remain a standard protoc/buf plugin
- Delete always means soft delete at the API layer
- Operations are opt-in — entities without `operations` produce no service output

## Out of Scope

- Go code generation (handlers, clients)
- Hard delete / purge RPCs
- gRPC or Connect runtime wiring
- Pagination cursor implementation details
- Databases or SQL — handled by existing modes
