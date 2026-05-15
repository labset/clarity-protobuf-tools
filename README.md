## clarity-protobuf-tools

## Development

### requirements

- [mise](https://mise.jdx.dev/) — manages `go`, `buf`, `golangci-lint` and `goreleaser` versions (see `.mise.toml`)

```bash
mise install
```

### housekeeping tasks

- install dependencies

```bash
mise run deps
```

- generate code from proto files

```bash
mise run generate
```

- lint the protos

```bash
mise run buf:lint
```

- format the protos

```bash
mise run buf:format
```

- lint the Go code

```bash
mise run lint
```

- build the binaries

```bash
mise run build
```

- install it locally

```bash
mise run install
```
