# clarity-protobuf-tools

A toolchain for working with [Clarity](https://github.com/labset)-annotated protobuf messages. It provides:

- **clarity-lint-plugin** — a buf lint plugin that validates proto messages annotated with clarity options
- **protoc-gen-clarity** — a protoc code generation plugin with multiple modes

## Quick Start

```bash
# install mise (manages go, buf, golangci-lint, goreleaser)
# see https://mise.jdx.dev/getting-started.html

mise install           # install tool versions
mise run deps          # download Go module dependencies
mise run generate      # generate Go code from proto files
mise run build         # build binaries (includes tests)
mise run install       # install plugins locally
```

### Dev Loop

```bash
# edit proto annotations or codegen templates, then:
mise run go:test       # run unit tests only
mise run go:vet        # run go vet
mise run lint          # lint Go + proto files
mise run lint:fix      # auto-fix lint issues

# e2e tests (requires Docker for testcontainers)
mise run e2e                        # run all e2e tests (generates + tests)
mise run e2e:connect-crud-outbox    # run connect-crud-outbox e2e only
mise run e2e:connect-handlers       # run connect-handlers e2e only
mise run e2e:mcp-tools              # run mcp-tools e2e only

# full build (unit tests + e2e + goreleaser)
mise run build
```

## Proto Annotations

Messages are annotated using `clarity.plugin.v1.ClarityMessageOptions`:

```protobuf
import "clarity/plugin/v1/options.proto";
import "clarity/plugin/v1/entity.proto";

message Product {
  option (clarity.plugin.v1.message) = {
    role: ROLE_ENTITY
    operations: [OPERATION_CREATE, OPERATION_GET, OPERATION_LIST, OPERATION_UPDATE, OPERATION_DELETE]
  };

  clarity.plugin.v1.Entity entity = 1;
  string name = 2;
  int64 price = 3;
}
```

### Roles

| Role | Description |
|------|-------------|
| `ROLE_ENTITY` | Marks a message as a database entity — must have an `entity` field of type `clarity.plugin.v1.Entity` at field number 1 |
| `ROLE_REFERENCE` | Marks a message as a reference type — must be defined in `refs.proto` with a `string id` field |

### Operations

| Operation | Description |
|-----------|-------------|
| `OPERATION_CREATE` | Generate a create RPC with typed `item` request |
| `OPERATION_GET` | Generate a get-by-id RPC with `buf.validate` UUID constraint |
| `OPERATION_LIST` | Generate a list RPC with cursor-based pagination |
| `OPERATION_UPDATE` | Generate an update RPC with `item` + `FieldMask` for partial updates |
| `OPERATION_DELETE` | Generate a soft-delete RPC with validated id |

Operations are opt-in — only specified operations produce service output.

### Field Annotations

Fields on entity messages can be annotated with `clarity.plugin.v1.ClarityFieldOptions`:

```protobuf
import "clarity/plugin/v1/options.proto";

message Product {
  option (clarity.plugin.v1.message) = {role: ROLE_ENTITY};
  clarity.plugin.v1.Entity entity = 1;
  CategoryRef category = 2 [(clarity.plugin.v1.field) = {foreign_key: true}];
  SupplierRef supplier = 3; // no FK — different store
  string name = 4;
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `foreign_key` | `bool` | `false` | When `true`, generates a `REFERENCES <table>(id)` constraint on the `_id` column |

### Reference Types

Reference types are defined in `refs.proto` files alongside `models.proto`:

```protobuf
// refs.proto
import "clarity/plugin/v1/options.proto";

message CategoryRef {
  option (clarity.plugin.v1.message) = {role: ROLE_REFERENCE};
  string id = 1;
}
```

When an entity field uses a ref type:
- **Schema**: produces a `<name>_id UUID NOT NULL` column (with optional `REFERENCES` constraint)
- **Queries**: includes the `_id` column in all CRUD queries
- **Mapper**: converts between `<Ref>Ref{Id: ...}` and `uuid.UUID`

## Lint Plugin

The `clarity-lint-plugin` is a [buf lint plugin](https://buf.build/docs/lint/plugins) that validates clarity-annotated proto messages. Install it locally, then configure it in your `buf.yaml`:

```bash
# install the plugin (must be on your PATH)
go install github.com/labset/clarity-protobuf-tools/cmd/clarity-lint-plugin@latest
```

```yaml
# buf.yaml
version: v2

modules:
  - path: protos

deps:
  - buf.build/labset/clarity-protobuf-tools

lint:
  use:
    - STANDARD
  plugins:
    - plugin: clarity-lint-plugin
```

Run linting with:

```bash
buf lint
```

### Rules

| Rule | Description |
|------|-------------|
| `CLARITY_ENTITY_FILE` | `ROLE_ENTITY` messages must be in files named `models.proto` under a `<provider>.<domain>.<version>` package |
| `CLARITY_ENTITY_FIELD` | `ROLE_ENTITY` messages must have an `entity` field of type `clarity.plugin.v1.Entity` at field number 1 |
| `CLARITY_REF_MESSAGE` | Every `ROLE_ENTITY` in `models.proto` must have a corresponding `<Model>Ref` with `ROLE_REFERENCE` in `refs.proto` |

All rules are enabled by default.

## Codegen Modes

### sqlc

Generates PostgreSQL schema, [sqlc](https://sqlc.dev/) queries, and sqlc config from `ROLE_ENTITY` messages.

```
protoc --clarity_out=. --clarity_opt=mode=sqlc proto/*.proto
```

For a message in package `acme.inventory.v1`, generates:

```
internal/acme/inventory/v1/
├── sql/
│   ├── schema.sql        # CREATE SCHEMA + CREATE TABLE
│   └── queries/
│       └── product.sql   # CRUD queries (get, list, create, update, delete, soft_delete)
└── sqlc.yaml             # sqlc config targeting PostgreSQL with pgx
```

### atlas-sqlc

Extends `sqlc` mode with [Atlas](https://atlasgo.io/) migration configuration.

```
protoc --clarity_out=. --clarity_opt=mode=atlas-sqlc proto/*.proto
```

Generates everything from `sqlc` plus:

```
internal/acme/inventory/v1/
├── atlas.hcl             # Atlas project config (schema source, migration dir, dev DB)
└── sql/
    └── baseline.sql      # Prevents Atlas from dropping the public schema
```

### service

Generates `.proto` service definitions from `ROLE_ENTITY` messages with opted-in operations.

```
protoc --clarity_out=. --clarity_opt=mode=service proto/*.proto
```

For a message in package `acme.inventory.v1` with all operations, generates:

```
acme/inventory/v1/
├── service_product.proto       # service ProductService { ... }
├── rpc_create_product.proto    # CreateProductRequest/Response
├── rpc_get_product.proto       # GetProductRequest/Response
├── rpc_list_product.proto      # ListProductsRequest/Response
├── rpc_update_product.proto    # UpdateProductRequest/Response
└── rpc_delete_product.proto    # DeleteProductRequest/Response
```

Generated services follow these conventions:

- **Create/Get/Update** responses use a typed `item` field
- **List** response uses `repeated items` with `next_page_token` for cursor pagination
- **Update** request includes `FieldMask` for partial updates
- **Get/Update/Delete** requests validate `id` with `buf.validate` UUID constraint
- **Delete** is always soft delete at the API layer

### connect-handlers

Generates Go [Connect-RPC](https://connectrpc.com/) handler scaffolding with stub implementations returning `Unimplemented` for any proto service definition. Unlike other modes, this processes services rather than entities and is standalone — no SQLC or Atlas composition.

```
protoc --clarity_out=. --clarity_opt=mode=connect-handlers proto/*.proto
```

For a service `ProductService` in package `acme.inventory.v1` with RPCs `CreateProduct`, `GetProduct`, `ListProducts`, generates:

```
internal/acme/inventory/v1/
└── api/
    ├── handler_product_service.go   # ProductServiceDeps, constructor, Connect registration
    ├── rpc_create_product.go        # stub returning CodeUnimplemented
    ├── rpc_get_product.go           # stub returning CodeUnimplemented
    └── rpc_list_products.go         # stub returning CodeUnimplemented
```

Generated files are scaffolding — they are only emitted if they do not already exist on disk, so developers can safely edit them without regeneration overwriting their changes.

### connect-crud

Generates Go [Connect-RPC](https://connectrpc.com/) handler implementations from `ROLE_ENTITY` messages, backed by SQLC-generated stores. This mode includes `atlas-sqlc` under the hood, so a single invocation produces the full stack: SQL schema, SQLC queries/config, Atlas migration config, and Connect handler code.

```
protoc --clarity_out=. --clarity_opt=mode=connect-crud,go_module=github.com/acme/app proto/*.proto
```

Requires the `go_module` parameter to derive the SQLC store import path.

For a message in package `acme.inventory.v1` with all operations, generates:

```
internal/acme/inventory/v1/
├── sql/                              # from atlas-sqlc (included automatically)
│   ├── schema.sql
│   ├── baseline.sql
│   └── queries/
│       └── product.sql
├── sqlc.yaml
├── atlas.hcl
└── api/                              # connect-crud handlers
    ├── handler_product.go         # ProductDeps, constructor, Connect service registration
    ├── mapper_product.go          # proto ↔ SQLC conversion functions
    ├── rpc_create_product.go      # Create with duplicate detection (CodeAlreadyExists)
    ├── rpc_get_product.go         # Get by ID with CodeNotFound
    ├── rpc_list_product.go        # Cursor-based pagination (page_size, page_token)
    ├── rpc_update_product.go      # Partial update via update_mask
    └── rpc_delete_product.go      # Soft delete with CodeNotFound
```

Generated handlers:

- Accept `...connect.HandlerOption` for interceptor injection
- Create the SQLC store once in the constructor from `*pgxpool.Pool`
- Use consistent Connect error codes: `CodeNotFound`, `CodeInvalidArgument`, `CodeAlreadyExists`

### Type Mapping (sqlc/atlas-sqlc)

| Proto Type | PostgreSQL Type |
|------------|----------------|
| `string` | `TEXT` |
| `bytes` | `BYTEA` |
| `bool` | `BOOLEAN` |
| `int32`, `sint32`, `sfixed32`, `uint32`, `fixed32` | `INTEGER` |
| `int64`, `sint64`, `sfixed64`, `uint64`, `fixed64` | `BIGINT` |
| `float` | `REAL` |
| `double` | `DOUBLE PRECISION` |
| `google.protobuf.Timestamp` | `TIMESTAMPTZ` |
| `google.protobuf.Duration` | `INTERVAL` |
| `google.protobuf.Struct` / `Value` | `JSONB` |
| `enum` | `TEXT` with `CHECK` constraint |
| `repeated <scalar>` | Array (e.g. `TEXT[]`) |
| `repeated <message>`, nested message, `map` | `JSONB` |
| `oneof` | Nullable columns per variant |
| `<Model>Ref` (ROLE_REFERENCE) | `UUID` (with optional `REFERENCES` via `foreign_key`) |

## Usage with Buf

### buf.gen.yaml

#### sqlc mode

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: internal
    opt:
      - mode=sqlc
```

#### atlas-sqlc mode

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: internal
    opt:
      - mode=atlas-sqlc
```

#### connect-handlers mode

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: .
    opt:
      - mode=connect-handlers
```

#### connect-crud mode

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: .
    opt:
      - mode=connect-crud
      - go_module=github.com/acme/app
```

#### mcp-tools mode

Generates MCP tool wrappers that invoke connect-crud handlers in-process. Includes `connect-crud` under the hood.

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: .
    opt:
      - mode=mcp-tools
      - go_module=github.com/acme/app
```

#### service mode

Generates `.proto` service definitions from entity messages — output alongside your source protos:

```yaml
version: v2
inputs:
  - directory: protos

plugins:
  - local: protoc-gen-clarity
    out: protos
    opt:
      - mode=service
```

### Running generation

```bash
# generate with a specific config
buf generate --template buf.gen.yaml

# or with multiple configs in sequence
buf generate --template buf.gen.yaml
buf generate --template buf.gen.services.yaml
```

## Development

### Requirements

- [mise](https://mise.jdx.dev/) — manages `go`, `buf`, `golangci-lint` and `goreleaser` versions (see `.mise.toml`)

```bash
mise install
```

### Tasks

```bash
mise run deps          # download Go module dependencies
mise run generate      # generate Go code from proto files
mise run build         # full build (unit tests + e2e + goreleaser)
mise run install       # install plugins locally
mise run lint          # lint Go + proto files
mise run lint:fix      # auto-fix lint issues
mise run format        # format Go + proto files
mise run buf:lint      # lint proto files only
mise run buf:format    # format proto files only
mise run go:lint       # lint Go code only
mise run go:format     # format Go code only
mise run go:test       # run unit tests
mise run go:vet        # run go vet
mise run e2e           # run all e2e tests
mise run e2e:connect-crud-outbox  # e2e for connect-crud-outbox
mise run e2e:connect-handlers     # e2e for connect-handlers
mise run e2e:mcp-tools            # e2e for mcp-tools
```
