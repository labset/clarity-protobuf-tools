# Requirements: MCP Tools Codegen

> Generate MCP tool definitions that wrap connect-crud handlers for in-process invocation using the official Go MCP SDK.

## Context

The clarity protobuf tools project already generates connect-crud handlers from annotated proto messages. To expose these same CRUD operations as MCP tools (for LLM-driven workflows), a new codegen mode is needed that reuses the connect-crud handler layer and generates MCP tool wrappers on top of it, using `github.com/modelcontextprotocol/go-sdk`.

## Requirements

### Functional

- [ ] FR-1: New `mcp-tools` codegen mode that invokes the `connect-crud` generator under the hood
- [ ] FR-2: Generate one MCP tool per CRUD operation annotated on the proto message (create, get, list, update, delete)
- [ ] FR-3: Generate typed input/output structs per tool operation, derived from the proto request/response messages
- [ ] FR-4: Generate tool handler functions that construct a `connect.Request`, invoke the connect-crud handler method in-process, and return the result
- [ ] FR-5: Generate a registration function that adds all tools for an entity to an `*mcp.Server` via `mcp.AddTool`
- [ ] FR-6: Tool names should follow a consistent convention (e.g. `create_product`, `get_product`, `list_products`, `update_product`, `delete_product`)

### Non-Functional

- [ ] NFR-1: Follow existing codegen conventions — `generator_` file prefix, embedded Go templates in `templates/mcp-tools/`, golden file tests in `testdata/golden/mcp-tools/`
- [ ] NFR-2: Templates must use Go `text/template`, no string concatenation

### Deferred

- [ ] DFR-1: MCP server entrypoint / transport wiring (separate spec)
- [ ] DFR-2: MCP resource generation (e.g. exposing entities as MCP resources)
- [ ] DFR-3: Support for `connect-crud-outbox` composition (outbox-aware MCP tools)

## Constraints

- Must use `github.com/modelcontextprotocol/go-sdk` (official SDK)
- Must compose with the existing `connect-crud` generator, not duplicate its logic

## Out of Scope

- Transport selection (stdio, SSE, etc.)
- Server `main.go` generation
- Authentication / authorization in MCP tools
- MCP prompts or resources
