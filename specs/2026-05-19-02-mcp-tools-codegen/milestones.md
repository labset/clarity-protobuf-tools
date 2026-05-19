# Milestones: MCP Tools Codegen

## M1: Generator scaffold and connect-crud composition

- **Status:** pending
- **Description:** Wire up the mcp-tools generator that delegates to connect-crud, and register the new mode in the codegen plugin.
- **Acceptance Criteria:**
  - [ ] `generator_mcp_tools.go` exists and calls the connect-crud generator
  - [ ] `mcp-tools` mode is registered and selectable in the codegen plugin
  - [ ] Running with `mcp-tools` mode still produces all connect-crud output

## M2: MCP tool templates and generation

- **Status:** pending
- **Description:** Create Go templates that generate typed input/output structs, tool handler functions, and a registration function for each CRUD operation.
- **Acceptance Criteria:**
  - [ ] Templates exist in `templates/mcp-tools/` for tool handlers and registration
  - [ ] One MCP tool generated per annotated CRUD operation
  - [ ] Tool names follow `<operation>_<entity>` convention
  - [ ] Registration function adds all tools to an `*mcp.Server` via `mcp.AddTool`
  - [ ] Golden file tests pass in `testdata/golden/mcp-tools/`

## M3: End-to-end validation

- **Status:** pending
- **Description:** E2E test using a proto fixture that generates the full mcp-tools output including underlying connect-crud files, verifying the generated code compiles and tools register correctly.
- **Acceptance Criteria:**
  - [ ] E2E test generates mcp-tools output from a proto fixture
  - [ ] Generated code compiles successfully
  - [ ] Tools can be registered on an `*mcp.Server` instance
