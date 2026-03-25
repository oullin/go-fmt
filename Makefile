SHELL := /bin/bash
.DEFAULT_GOAL := help

APP := go-fmt
CMD := ./cmd/fmt
GO_WORKDIR := semantic
BUILD_DIR := bin
BIN := $(BUILD_DIR)/$(APP)
OXFMT_BIN := tooling/node_modules/.bin/oxfmt

ARGS ?= . ## Files or directories to target
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev) ## Build version injected into binaries
CGO_ENABLED ?= 0 ## CGO setting used for build and release
DIST_DIR ?= dist ## Directory for release binaries
RELEASE_PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 ## Space-separated GOOS/GOARCH release targets

.PHONY: help format build release test test-race test-short vet fmt-source install clean

help: ## Show available targets and override variables
	@awk '\
		BEGIN { \
			FS = "## "; \
		} \
		function trim(value) { \
			gsub(/^[[:space:]]+|[[:space:]]+$$/, "", value); \
			return value; \
		} \
		/^[a-zA-Z0-9_.-]+:.*## / { \
			split($$1, parts, ":"); \
			targets[++target_count] = sprintf("  %-14s %s", parts[1], $$2); \
			next; \
		} \
		/^[A-Z_][A-Z0-9_]*[[:space:]]*\?*=.*## / { \
			split($$1, parts, "\\?="); \
			vars[++var_count] = sprintf("  %-18s %-24s %s", trim(parts[1]), trim(parts[2]), $$2); \
		} \
		END { \
			printf "go-fmt developer targets\n\nTargets:\n"; \
			for (i = 1; i <= target_count; i++) { \
				print targets[i]; \
			} \
			if (var_count) { \
				printf "\nVariables:\n"; \
				for (i = 1; i <= var_count; i++) { \
					print vars[i]; \
				} \
			} \
		} \
	' $(MAKEFILE_LIST)
	@printf "\nExamples:\n"
	@printf "  pnpm turbo run check --filter=semantic\n"
	@printf "  pnpm turbo run check --filter=tooling\n"
	@printf "  make fmt-source\n"

format: ## Apply formatter changes to ARGS
	@OXFMT_BIN="$(OXFMT_BIN)" ./scripts/format.sh $(ARGS)

build: ## Build a host-native binary into ./bin
	@VERSION='$(strip $(VERSION))' ./scripts/build.sh

release: ## Build release binaries into $(DIST_DIR)
	@VERSION='$(strip $(VERSION))' DIST_DIR='$(strip $(DIST_DIR))' RELEASE_PLATFORMS='$(strip $(RELEASE_PLATFORMS))' ./scripts/release.sh

test: ## Run all tests with verbose output
	pnpm test

test-race: ## Run all tests with the race detector
	go -C $(GO_WORKDIR) test ./... -race -v

test-short: ## Run tests in short mode
	go -C $(GO_WORKDIR) test ./... -short

vet: ## Run go vet across the module
	pnpm vet

fmt-source: ## Rewrite Go source formatting in the repository
	@./scripts/fmt-source.sh

install: ## Install the CLI with go install
	go -C $(GO_WORKDIR) install $(CMD)

clean: ## Remove build artifacts and clean the Go cache
	rm -f $(BIN)
	rm -rf $(DIST_DIR)
	rm -rf .turbo node_modules tooling/node_modules semantic/node_modules
	go -C $(GO_WORKDIR) clean -cache
