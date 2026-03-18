# go-fmt

`go-fmt` is a Go coding styles and standards fixer for code produced by humans, agents, and other automations.

The repo exposes:

- a reusable engine for checking and formatting Go source files
- a CLI named `fmt`
- a rule-based style pipeline instead of preset bundles

## Current Scope

The first built-in semantic rule is `spacing`.

It enforces spacing around control flow, `return`, `defer`, `var`, and type declarations, and it checks that type definitions stay at the beginning of the file.

Formatting runs in this order:

1. `spacing`
2. `gofmt`
3. `goimports`

## CLI

The binary is exposed as `fmt`.

```bash
fmt check [paths...]
fmt format [paths...]
```

Common flags:

```bash
fmt check --config ./go-fmt.yml --format text .
fmt check --format json .
fmt check --format agent .
fmt format .
```

Supported output formats:

- `text`
- `json`
- `agent`

Exit behavior:

- `fmt check` exits `0` when no changes are needed
- `fmt check` exits `1` when violations or errors are found
- `fmt format` exits `0` when formatting succeeds
- `fmt format` exits `1` when an error occurs

## Using `fmt` In This Repo

Run the CLI from source without installing it:

```bash
go run ./cmd/fmt --help
go run ./cmd/fmt check .
go run ./cmd/fmt format .
```

Run against a specific file:

```bash
go run ./cmd/fmt check ./internal/rules/spacing/spacing.go
go run ./cmd/fmt format ./internal/rules/spacing/spacing.go
```

Use the example config in this repo as a starting point:

```bash
cp go-fmt.yml.example go-fmt.yml
go run ./cmd/fmt check --config ./go-fmt.yml .
```

Build the local binary and run it from `./builds/fmt`:

```bash
make build
./builds/fmt check .
./builds/fmt format .
```

Make targets for day-to-day repo use:

```bash
make help
make config
make check
make format
make check ARGS=./internal/rules/spacing/spacing.go
make check-json
make check-agent
```

Optional variables:

- `ARGS`: files or directories to target, defaults to `.`
- `CONFIG`: config path, for example `make check CONFIG=./go-fmt.yml`
- `OUTPUT`: text output mode for `make check`, defaults to `text`

## Config

The default config file is `go-fmt.yml`.

Example:

```yaml
rules:
  spacing:
    enabled: true

formatters:
  gofmt: true
  goimports: true

exclude:
  - .git
  - vendor

not_path:
  - third_party/generated

not_name:
  - "*.pb.go"
```

Config fields:

- `rules.spacing.enabled`: enable or disable the spacing rule
- `formatters.gofmt`: run `gofmt` after rule-based formatting
- `formatters.goimports`: run `goimports` after `gofmt`
- `exclude`: directories to skip while walking trees
- `not_path`: path substrings to skip
- `not_name`: filename glob patterns to skip

If no config file exists, built-in defaults are used.

## File Discovery

The engine:

- accepts files or directories
- scans `.go` files
- skips hidden directories, `.git`, and `vendor`
- skips `*.gen.go`
- skips files starting with `// Code generated`

## Development

Build:

```bash
make build
```

Test:

```bash
make test
```

Install `goimports` locally if you want the optional formatter step available during CLI runs:

```bash
make install-tools
```

Install the CLI locally:

```bash
make install
```

## Package Layout

- `cmd/fmt`: CLI entrypoint
- `internal/config`: YAML config loading through Viper
- `internal/engine`: file collection, orchestration, reporting, formatter pipeline
- `internal/rules`: rule contracts
- `internal/rules/spacing`: first semantic style rule

## Notes

- This project is not a Go port of an existing Laravel formatter.
- The goal is to provide a Go-native style engine and fixer framework for automated code generation workflows.
- `goimports` is optional at runtime. If it is not installed, the engine skips that step.
