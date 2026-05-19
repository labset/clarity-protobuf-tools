# Milestones: MCP Tools Codegen

## M1: Generator scaffold and connect-crud composition

- **Status:** done
- **Description:** Wire up the mcp-tools generator that delegates to connect-crud, and register the new mode in the codegen plugin.
- **Acceptance Criteria:**
  - [x] `generator_mcp_tools.go` exists and calls the connect-crud generator
  - [x] `mcp-tools` mode is registered and selectable in the codegen plugin
  - [x] Running with `mcp-tools` mode still produces all connect-crud output

## M2: MCP tool templates and generation

- **Status:** done
- **Description:** Create Go templates that generate tool handler functions and a registration function for each CRUD operation.
- **Acceptance Criteria:**
  - [x] Templates exist in `templates/mcp-tools/` for tool handlers and registration
  - [x] One MCP tool generated per annotated CRUD operation
  - [x] Tool names follow `<operation>_<entity>` convention
  - [x] Registration function adds all tools to an `*mcp.Server` via `server.AddTool`
  - [x] Golden file tests pass in `testdata/golden/mcp-tools/`

## M3: End-to-end validation

- **Status:** done
- **Description:** E2E test using a proto fixture that generates the full mcp-tools output including underlying connect-crud files, verifying the generated code compiles and tools register correctly.
- **Acceptance Criteria:**
  - [x] E2E test generates mcp-tools output from a proto fixture
  - [x] Generated code compiles successfully
  - [x] Tools can be registered on an `*mcp.Server` instance
