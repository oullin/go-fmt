# go-fmt

**A semantic style engine and fixer for Go source code.**

`go-fmt` goes beyond `gofmt`. It applies rule-based semantic formatting — enforcing blank-line boundaries around control flow, ensuring type declarations appear at the top of the file, and normalising spacing around `var`, `defer`, and `return` — then finishes with `gofmt` and `goimports`. The result is consistently styled Go, whether the code was written by a person, an agent, or a code generator.

The project ships as a **reusable engine** and a **standalone CLI** (`fmt`). Rules run first, formatters run second — giving you deterministic, layered formatting in a single pass.

---

## Table of Contents

- [Quick Start](#quick-start)
- [CLI](#cli)
- [Configuration](#configuration)
- [Spacing Rule](#spacing-rule)
- [File Discovery](#file-discovery)
- [Output Formats](#output-formats)
- [Exit Codes](#exit-codes)
- [Development](#development)
- [Package Layout](#package-layout)

---

## Quick Start

**Install the CLI:**

```bash
go install github.com/oullin/go-fmt/cmd/fmt@latest
```

**Or build from source:**

```bash
make build
./bin/fmt check .
./bin/fmt format .
```

If you downloaded a standalone binary from an older release, make sure it is executable before running it:

```bash
chmod +x /path/to/fmt
```

On macOS, older unsigned release binaries may be quarantined by Gatekeeper. If the binary exits immediately with `killed`, clear the quarantine attribute on that binary before running it:

```bash
xattr -d com.apple.quarantine /path/to/fmt
```

**Run directly without installing:**

```bash
go run ./cmd/fmt check .
go run ./cmd/fmt format .
```

---

## CLI

The binary is called `fmt` and has two commands: **check** and **format**.

```
fmt check  [paths...]   # report violations without writing changes
fmt format [paths...]   # fix violations and write changes to disc
```

Both commands accept these flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to a `go-fmt.yml` config file | auto-detected in working directory |
| `--format` | Output format: `text`, `json`, or `agent` | `text` |

### Examples

```bash
# check everything in the current directory
fmt check .

# check with a specific config and JSON output
fmt check --config ./go-fmt.yml --format json .

# format a single file
fmt format ./rules/spacing/spacing.go

# agent-friendly output for CI integrations
fmt check --format agent .
```

---

## Configuration

`go-fmt` looks for a `go-fmt.yml` file in the working directory. If none is found, built-in defaults apply. You can also point to a config explicitly with `--config`.

**Copy the example config to get started:**

```bash
cp go-fmt.yml.example go-fmt.yml
```

### Full Config Reference

```yaml
# Enable or disable individual semantic rules.
rules:
  spacing:
    enabled: true        # enforce blank-line spacing (default: true)

# Enable or disable post-rule formatters.
formatters:
  gofmt: true            # run gofmt after rules (default: true)
  goimports: true        # run goimports after gofmt (default: true)

# Directories to skip entirely during file walking.
exclude:
  - .git
  - vendor

# Path substrings — any file whose path contains a match is skipped.
not_path:
  - third_party/generated

# Filename glob patterns — any file whose name matches is skipped.
not_name:
  - "*.pb.go"
```

| Field | Type | Description |
|-------|------|-------------|
| `rules.spacing.enabled` | bool | Toggle the spacing rule on or off |
| `formatters.gofmt` | bool | Run `gofmt` after semantic rules |
| `formatters.goimports` | bool | Run `goimports` after `gofmt` |
| `exclude` | list | Directory names to skip during tree walking |
| `not_path` | list | Substring matches against full file paths |
| `not_name` | list | Glob patterns matched against file names |

> **Note:** `goimports` is optional at runtime. If it is not installed on the system, the engine skips that step silently — no error is raised.

---

## Spacing Rule

The `spacing` rule is the first built-in semantic rule. It inspects Go source files using the AST and enforces consistent blank-line boundaries.

### What It Enforces

**Blank line _before_ control flow and keywords:**

A blank line is required before `if`, `for`, `range`, `switch`, `select`, `defer`, `return`, `continue`, `break`, `goto`, and `fallthrough` — when they are not the first statement in a block.

```go
// before
func run() {
    x := 1
    if x > 0 {
        println("positive")
    }
    return
}

// after
func run() {
    x := 1

    if x > 0 {
        println("positive")
    }

    return
}
```

**Blank line _after_ block statements:**

A blank line is required after `if`, `for`, `range`, `switch`, `select`, and `defer` blocks when followed by another statement.

```go
// before
func run() {
    if ready {
        start()
    }
    cleanup()
}

// after
func run() {
    if ready {
        start()
    }

    cleanup()
}
```

**Blank lines around `var` declarations:**

A blank line is required before and after standalone `var` declarations — unless the preceding or following statement is also a `var` or short assignment (`:=`), in which case they are allowed to stay grouped.

```go
// before
func run() {
    x := setup()
    var cfg Config
    process(cfg)
}

// after
func run() {
    x := setup()

    var cfg Config

    process(cfg)
}
```

**Type declarations at the top of the file:**

All `type` definitions must appear before any non-type declarations (after the import block). If types are scattered throughout the file, the rule reorders them to the top and reports a violation.

**Blank lines around type declarations:**

Blank lines are required before and after `type` declarations to visually separate them from surrounding code.

### Where It Applies

The spacing rule inspects statement lists inside:

- Function bodies (`BlockStmt`)
- `case` and `default` clauses (`CaseClause`)
- `select` communication clauses (`CommClause`)

---

## File Discovery

When given directories, the engine walks them recursively and collects `.go` files. The following are always skipped:

| Skipped | Reason |
|---------|--------|
| Hidden directories (`.foo/`) | Convention — not source code |
| `.git/` | Repository metadata |
| `vendor/` | Vendored dependencies |
| `*.gen.go` files | Generated code by convention |
| Files starting with `// Code generated` | Go standard for generated files |
| Paths matching `exclude` config | User-defined directory exclusions |
| Paths matching `not_path` config | User-defined path substring exclusions |
| Files matching `not_name` config | User-defined filename glob exclusions |

When given individual files, they are used directly (filtering still applies).

If no paths are provided, the engine defaults to the current directory (`.`).

---

## Output Formats

### Text (default)

Human-readable output with relative file paths, violation details, and a summary line.

```
~ engine/engine.go:42 [spacing] missing blank line before if
  would apply spacing
Result: fail. 1 changed, 1 violation(s), 0 error(s).
```

### JSON

Structured output with full details for each file. Useful for editors, dashboards, and tooling.

```json
{
  "result": "fail",
  "files": 1,
  "changed": 1,
  "results": [
    {
      "file": "engine.go",
      "applied": ["spacing"],
      "violations": [
        {
          "rule": "spacing",
          "line": 42,
          "message": "missing blank line before if"
        }
      ],
      "changed": true
    }
  ]
}
```

### Agent

Compact JSON designed for AI agents and CI pipelines. Groups output by changed files and violations rather than per-file results.

```json
{
  "result": "fail",
  "summary": { "files": 1, "changed": 1, "violations": 1 },
  "changed": [
    { "file": "engine.go", "steps": ["spacing"] }
  ],
  "violations": [
    {
      "file": "engine.go",
      "rule": "spacing",
      "line": 42,
      "message": "missing blank line before if"
    }
  ]
}
```

---

## Exit Codes

| Command | Code | Meaning |
|---------|------|---------|
| `fmt check` | `0` | No violations found — code is clean |
| `fmt check` | `1` | Violations or errors detected |
| `fmt format` | `0` | Formatting applied successfully |
| `fmt format` | `1` | An error occurred during formatting |

---

## Development

### Prerequisites

- Go 1.24+
- `goimports` (optional, for the import formatting step)

### Make Targets

```bash
make help            # list all targets and variables
make build           # compile a host-native binary to ./bin/fmt
make release         # build release binaries to ./dist
make test            # run all tests with verbose output
make test-race       # run tests with race detector
make test-short      # run tests in short mode
make vet             # run go vet
make lint            # gofmt + test + vet
make install         # go install the CLI
make install-tools   # install goimports
make config          # copy go-fmt.yml.example to go-fmt.yml
make clean           # remove build artefacts and clean cache
```

### Make Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ARGS` | `.` | Files or directories to target |
| `CONFIG` | _(empty)_ | Path to config file |
| `OUTPUT` | `text` | Output format for check commands |

```bash
# check a specific file
make check ARGS=./rules/spacing/spacing.go

# use a specific config
make check CONFIG=./go-fmt.yml

# JSON output
make check-json

# agent output
make check-agent
```

### Release Signing

Browser-downloaded macOS binaries need Apple Developer ID signing and notarization to pass Gatekeeper cleanly. The release workflow expects these GitHub Actions secrets:

- `APPLE_KEYCHAIN_PASSWORD`
- `APPLE_CERTIFICATE_P12_BASE64`
- `APPLE_CERTIFICATE_PASSWORD`
- `APPLE_SIGNING_IDENTITY`
- `APPLE_ID`
- `APPLE_APP_SPECIFIC_PASSWORD`
- `APPLE_TEAM_ID`

The workflow signs `dist/fmt-darwin-arm64`, notarizes a ZIP containing that binary, packages the Linux build as `dist/fmt-linux-amd64.tar.gz`, and publishes those archives as the release assets so executable permissions are preserved on download. Apple does not support stapling tickets to standalone binaries, so the macOS binary inside the ZIP relies on Gatekeeper’s online ticket lookup at first launch.

---

## Package Layout

```
cmd/fmt/                   CLI entrypoint and output rendering
config/                    YAML config loading via Viper, with defaults
engine/                    File collection, rule orchestration, formatter pipeline, reporting
rules/                     Rule interface contract
rules/spacing/             Spacing semantic rule (AST-based)
```

### Formatting Pipeline

The engine processes each file through a layered pipeline:

```
source file --> spacing rule --> gofmt --> goimports --> result
```

Each step only runs if enabled in the config. If a file is unchanged after all steps, it is reported as clean.

---

## Notes

- This is a Go-native project — not a port of an existing formatter from another ecosystem.
- The rule system is designed to be extended. New rules implement the `Rule` interface (`Name()` + `Apply()`) and are registered in the engine.
- `goimports` is optional. If the binary is not found at runtime, that step is skipped without error.
