## go-protoc-gen-plugin

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=viqueen_protoc-gen-plugin&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=viqueen_protoc-gen-plugin)

Template repository for creating protoc plugins in Go.

### usage

- install it

````bash
go install github.com/<username>/go-protoc-gen-plugin/cmd@latest
````

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
