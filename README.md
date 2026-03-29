# go-fmt

**A semantic style engine and fixer for Go source code.**

`go-fmt` goes beyond `gofmt`. It applies rule-based semantic formatting — enforcing blank-line boundaries around control flow, ensuring type declarations appear at the top of the file, and normalising spacing around `var`, `defer`, and `return` — then finishes with `gofmt` and `goimports`. The result is consistently styled Go, whether the code was written by a person, an agent, or a code generator.

The project ships as a **reusable engine** and a **standalone CLI** (`go-fmt`). Rules run first; formatters run second — giving you deterministic, layered formatting in a single pass. The repository itself is a small Turborepo: the Go formatter lives in the `semantic` workspace, while the `tooling` workspace owns Oxc-based formatting for every supported non-Go file type.

---

## Table of Contents

- [Highlights](#highlights)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Running with Docker](#running-with-docker)
- [CLI](#cli)
- [Configuration](#configuration)
- [Spacing Rule](#spacing-rule)
- [File Discovery](#file-discovery)
- [Output Formats](#output-formats)
- [Exit Codes](#exit-codes)
- [Development](#development)
- [Package Layout](#package-layout)

---

## Highlights

- Adds semantic formatting on top of `gofmt`, not just whitespace normalisation
- Runs as either a local CLI, a Docker image, or directly from source
- Supports human-readable, JSON, and agent-oriented output
- Applies rule fixes first, then standard Go formatters for predictable results
- Can be embedded as a reusable engine in Go code

---

## Quick Start

The fastest way to try `go-fmt` with Docker — no local install required:

```bash
docker run --rm -v "$PWD":/work -w /work ghcr.io/oullin/go-fmt:latest check .
docker run --rm -v "$PWD":/work -w /work ghcr.io/oullin/go-fmt:latest format .
```

With Go installed locally:

```bash
go install github.com/oullin/go-fmt/semantic/cmd/fmt@latest
go-fmt check .
go-fmt format .
```

---

## Installation

### Install with Go

Requires Go 1.25 or newer.

```bash
go install github.com/oullin/go-fmt/semantic/cmd/fmt@latest
```

Make sure your Go bin directory is on `PATH` so the `go-fmt` binary is available in your shell:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Verify the installation:

```bash
go-fmt version
```

### Build Locally

Build a host-native binary into `./bin/go-fmt` from this repository:

```bash
make build
./bin/go-fmt version
```

To install from the local source tree into your Go bin directory:

```bash
make install
go-fmt version
```

### Run from Source

No installation needed — run directly from the repository:

```bash
go run ./semantic/cmd/fmt check .
go run ./semantic/cmd/fmt format .
```

---

## Running with Docker

The `go-fmt` image is published to `ghcr.io/oullin/go-fmt` and is publicly available — no authentication required. The image supports `linux/amd64` and `linux/arm64`.

On macOS, this runs as a Linux container via Docker Desktop.

### Direct Docker Run

Run the published container image against your current directory:

```bash
docker run --rm -v "$PWD":/work -w /work ghcr.io/oullin/go-fmt:latest check .
docker run --rm -v "$PWD":/work -w /work ghcr.io/oullin/go-fmt:latest format .
```

Running without a subcommand prints the help menu:

```bash
docker run --rm ghcr.io/oullin/go-fmt:latest
```

### Docker Compose

The recommended way to integrate `go-fmt` into a project is with a Compose file. This keeps the command short and the configuration reusable.

#### Adding go-fmt to Your Project

Download the example Compose file into your project:

```bash
curl -o go-fmt.compose.yaml https://raw.githubusercontent.com/oullin/go-fmt/main/examples/consumer/go-fmt.compose.yaml
```

The Compose file is intentionally minimal:

```yaml
services:
    go-fmt:
        image: ghcr.io/oullin/go-fmt:latest
        working_dir: /work
        volumes:
            - .:/work
        environment:
            HOST_PROJECT_PATH: ${PWD}
        command: ['help']
```

It mounts the caller's current directory to `/work` inside the container and sets `HOST_PROJECT_PATH` so `--host-path` can map host paths back to the mounted tree.

#### Running the Formatter

The default command is `help`. To see available commands and flags, run:

```bash
docker compose -f go-fmt.compose.yaml run --rm go-fmt
```

Override the default command with `check` or `format` to run the formatter:

```bash
docker compose -f go-fmt.compose.yaml run --rm go-fmt check .
docker compose -f go-fmt.compose.yaml run --rm go-fmt format .
```

The `--rm` flag removes the container automatically after it exits.

#### Using a Project-Local Compose File

If your project already includes a Compose file with a `go-fmt` service (for example, `api/docker.api.compose.yaml`), run it the same way:

```bash
# View available commands
docker compose -f api/docker.api.compose.yaml run --rm go-fmt

# Check for violations without writing changes
docker compose -f api/docker.api.compose.yaml run --rm go-fmt check .

# Fix violations and write changes to disk
docker compose -f api/docker.api.compose.yaml run --rm go-fmt format .
```

This spins up the `ghcr.io/oullin/go-fmt:latest` container, mounts your working directory into `/work`, runs the formatter against your codebase, and cleans up the container when done.

#### Sharing a Compose File Across Projects

If you keep the Compose file somewhere central and reuse it from other projects, pass `--project-directory "$PWD"` so the volume mount binds the current project instead of the directory that stores the Compose file:

```bash
docker compose -f /path/to/go-fmt.compose.yaml --project-directory "$PWD" run --rm go-fmt check .
docker compose -f /path/to/go-fmt.compose.yaml --project-directory "$PWD" run --rm go-fmt format .
```

To target a subdirectory with `--host-path`:

```bash
docker compose -f /path/to/go-fmt.compose.yaml --project-directory "$PWD" run --rm go-fmt format --host-path "$PWD/pkg/api"
```

Paths outside the caller's current directory are intentionally rejected.

---

## CLI

The binary is called `go-fmt` and exposes two primary commands.

| Command                    | Purpose                                   |
| -------------------------- | ----------------------------------------- |
| `go-fmt check [paths...]`  | report violations without writing changes |
| `go-fmt format [paths...]` | fix violations and write changes to disk  |

If no paths are provided, both commands default to the current directory (`.`).

Both commands accept the same flags:

| Flag          | Description                                                                          | Default                            |
| ------------- | ------------------------------------------------------------------------------------ | ---------------------------------- |
| `--config`    | Path to a `config.yml` config file                                                   | auto-detected in working directory |
| `--cwd`       | Path used for config discovery and report-relative file paths                        | current working directory          |
| `--format`    | Output format: `text`, `json`, or `agent`                                            | `text`                             |
| `--host-path` | Absolute host path under `HOST_PROJECT_PATH`; intended for the consumer Compose flow | _(disabled unless env is set)_     |

The standalone CLI formats Go source only. Repository-local `make format` also runs `oxfmt` across every supported non-Go file type in the repo, excluding `*.go`. CI checks run per workspace with `pnpm turbo run check --filter=semantic` and `pnpm turbo run check --filter=tooling`.

### Common Workflows

```bash
# check everything in the current directory
go-fmt check .

# check with a specific config and JSON output
go-fmt check --config ./config.yml --format json .

# check a host path mounted by the consumer Compose file
go-fmt check --host-path /absolute/host/project/pkg/api

# format a single file
go-fmt format ./semantic/rules/spacing/spacing.go

# format a host path mounted by the consumer Compose file
go-fmt format --host-path /absolute/host/project/pkg/api

# agent-friendly output for CI integrations
go-fmt check --format agent .
```

---

## Configuration

`go-fmt` looks for a `config.yml` file in the working directory. If none is found, built-in defaults apply. You can also point to a config explicitly with `--config path/to/config.yml`.

All fields are optional — only include the ones you want to override.

```yaml
# Enable or disable individual semantic rules.
rules:
    spacing:
        enabled: true # default: true

# Enable or disable post-rule formatters.
formatters:
    gofmt: true # default: true
    goimports: true # default: true

# Directories to skip entirely during file walking.
exclude:
    - .git # default
    - node_modules # default
    - vendor # default

# Path substrings — any file whose path contains a match is skipped.
not_path:
    - third_party/generated

# Filename glob patterns — any file whose name matches is skipped.
not_name:
    - '*.pb.go'
```

| Field                   | Type | Default                          | Description                                 |
| ----------------------- | ---- | -------------------------------- | ------------------------------------------- |
| `rules.spacing.enabled` | bool | `true`                           | Toggle the spacing rule on or off           |
| `formatters.gofmt`      | bool | `true`                           | Run `gofmt` after semantic rules            |
| `formatters.goimports`  | bool | `true`                           | Run `goimports` after `gofmt`               |
| `exclude`               | list | `.git`, `node_modules`, `vendor` | Directory names to skip during tree walking |
| `not_path`              | list | _(empty)_                        | Substring matches against full file paths   |
| `not_name`              | list | _(empty)_                        | Glob patterns matched against file names    |

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

| Skipped                                 | Reason                                 |
| --------------------------------------- | -------------------------------------- |
| Hidden directories (`.foo/`)            | Convention — not source code           |
| `.git/`                                 | Repository metadata                    |
| `vendor/`                               | Vendored dependencies                  |
| `*.gen.go` files                        | Generated code by convention           |
| Files starting with `// Code generated` | Go standard for generated files        |
| Paths matching `exclude` config         | User-defined directory exclusions      |
| Paths matching `not_path` config        | User-defined path substring exclusions |
| Files matching `not_name` config        | User-defined filename glob exclusions  |

When given individual files, they are used directly (filtering still applies).

If no paths are provided, the engine defaults to the current directory (`.`).

---

## Output Formats

### Text (default)

Human-readable output with relative file paths, violation details, and a summary line.

```
~ semantic/engine/engine.go:42 [spacing] missing blank line before if
  would apply spacing
Result: fail. 1 changed, 1 violation(s), 0 error(s).
```

### JSON

Structured output with full details for each file. Useful for editors, dashboards, and scripts.

```json
{
	"result": "fail",
	"files": 1,
	"changed": 1,
	"results": [
		{
			"file": "semantic/engine/engine.go",
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

Compact JSON designed for AI agents and CI pipelines. It groups changed files and violations rather than mirroring every per-file report field.

```json
{
	"result": "fail",
	"summary": { "files": 1, "changed": 1, "violations": 1 },
	"changed": [{ "file": "semantic/engine/engine.go", "steps": ["spacing"] }],
	"violations": [
		{
			"file": "semantic/engine/engine.go",
			"rule": "spacing",
			"line": 42,
			"message": "missing blank line before if"
		}
	]
}
```

---

## Exit Codes

| Command         | Code | Meaning                             |
| --------------- | ---- | ----------------------------------- |
| `go-fmt check`  | `0`  | No violations found — code is clean |
| `go-fmt check`  | `1`  | Violations or errors detected       |
| `go-fmt format` | `0`  | Formatting applied successfully     |
| `go-fmt format` | `1`  | An error occurred during formatting |

---

## Development

### Prerequisites

- Go 1.25 or newer
- Node.js 25.8.2 with `pnpm` 10.32.1 for repo-local `make` and Turbo workflows
- Docker Desktop or another Docker runtime if you use the published container image

Install workspace dependencies before using the repo-local tasks:

```bash
nvm use
pnpm install
```

For local day-to-day development, `make help` lists the maintained task entrypoints.

### Make Targets

```bash
make help            # list all targets and variables
pnpm turbo run check --filter=semantic   # run semantic workspace checks
pnpm turbo run check --filter=tooling    # run tooling workspace checks
make format          # apply Go fixes plus Oxc formatting for supported non-Go files
make build           # compile a host-native binary to ./bin/go-fmt
make release         # build release binaries into ./dist
make test            # run all tests with verbose output
make test-race       # run tests with race detector
make test-short      # run tests in short mode
make vet             # run go vet
make fmt-source      # rewrite Go source formatting in the repo
make install         # go install the CLI
make clean           # remove build artefacts and clean cache
```

### Make Variables

| Variable            | Default                                             | Description                                          |
| ------------------- | --------------------------------------------------- | ---------------------------------------------------- |
| `ARGS`              | `.`                                                 | Files or directories to target                       |
| `VERSION`           | `git describe ...` or `dev`                         | Build version injected into binaries                 |
| `CGO_ENABLED`       | `0`                                                 | CGO setting for build and release                    |
| `DIST_DIR`          | `dist`                                              | Directory for release binaries                       |
| `RELEASE_PLATFORMS` | `darwin/amd64 darwin/arm64 linux/amd64 linux/arm64` | Space-separated GOOS/GOARCH pairs for `make release` |

```bash
# run semantic workspace checks
pnpm turbo run check --filter=semantic

# run tooling-only checks
pnpm turbo run check --filter=tooling

# format a repo-root Markdown file
make format ARGS=README.md

# rewrite Go source formatting
make fmt-source

# override release output directory
make release DIST_DIR=builds
```

### Docker Distribution

The published package for `go-fmt` is the multi-arch container image `ghcr.io/oullin/go-fmt`. The image is public — no authentication is required to pull it. The release workflow:

- runs tests on every release build
- creates the next Git tag and GitHub release
- publishes `ghcr.io/oullin/go-fmt:latest`
- publishes `ghcr.io/oullin/go-fmt:<tag>`
- pushes a multi-arch image for `linux/amd64` and `linux/arm64`

---

## Package Layout

```
semantic/cmd/fmt/          Standalone Go CLI entrypoint
semantic/config/           YAML config loading via Viper, with defaults
semantic/engine/           Go file collection, rule orchestration, formatter pipeline, reporting
semantic/rules/            Rule interface contract
semantic/rules/spacing/    Spacing semantic rule (AST-based)
tooling/                   Oxc-based formatting for supported non-Go file types
```

### Formatting Pipeline

The engine processes each file through a layered pipeline:

```
source file --> spacing rule --> gofmt --> goimports --> result
```

Each step only runs if enabled in the config. If a file is unchanged after all steps, it is reported as clean.

---

## Notes

- The standalone formatter remains Go-native.
- Repo-local Make targets extend that with Oxc, so supported non-Go files are kept in sync without maintaining an extension allowlist.
- The rule system is designed to be extended. New rules implement the `Rule` interface (`Name()` + `Apply()`) and are registered in the engine.
