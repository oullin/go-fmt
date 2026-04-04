SHELL := /bin/bash
.DEFAULT_GOAL := help

APP := go-fmt
CMD := ./cmd/fmt
GO_WORKDIR := semantic
BUILD_DIR := bin
BIN := $(BUILD_DIR)/$(APP)
OXFMT_BIN := tooling/node_modules/.bin/oxfmt

ARGS ?= . ## With '.', format changed tracked and untracked files; set a path to target a specific subtree
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev) ## Build version injected into binaries
CGO_ENABLED ?= 0 ## CGO setting used for build and release
DIST_DIR ?= dist ## Directory for release binaries
RELEASE_PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 ## Space-separated GOOS/GOARCH release targets

.PHONY: help format build release test test-race test-short vet fmt-source install clean

help: ## Show available targets and override variables
	@# Parse Make metadata and render styled help output through the dedicated helper script.
	@./scripts/help.sh $(MAKEFILE_LIST)

format: ## Apply formatter changes to ARGS
	@# Run the repository formatter script against the requested files or directories.
	@OXFMT_BIN="$(OXFMT_BIN)" ./scripts/format.sh $(ARGS)

build: ## Build a host-native binary into ./bin
	@# Compile the current version into a local binary for the host platform.
	@VERSION='$(strip $(VERSION))' ./scripts/build.sh

release: ## Build release binaries into $(DIST_DIR)
	@# Produce distributable binaries for every configured GOOS/GOARCH target.
	@VERSION='$(strip $(VERSION))' DIST_DIR='$(strip $(DIST_DIR))' RELEASE_PLATFORMS='$(strip $(RELEASE_PLATFORMS))' ./scripts/release.sh

test: ## Run all tests with verbose output
	@# Execute the workspace test suite through the package-level test script.
	pnpm test

test-race: ## Run all tests with the race detector
	@# Run Go tests with race detection enabled for concurrency-sensitive changes.
	go -C $(GO_WORKDIR) test ./... -race -v

test-short: ## Run tests in short mode
	@# Run the fast Go test subset intended for quick local verification.
	go -C $(GO_WORKDIR) test ./... -short

vet: ## Run go vet across the module
	@# Run static analysis checks configured for the repository workspace.
	pnpm vet

fmt-source: ## Rewrite Go source formatting in the repository
	@# Normalize Go source formatting across tracked repository files.
	@./scripts/fmt-source.sh

install: ## Install the CLI with go install
	@# Install the CLI into the active Go bin directory for local use.
	go -C $(GO_WORKDIR) install $(CMD)

clean: ## Remove build artifacts and clean the Go cache
	@# Remove the host-native build artifact from the local bin directory.
	rm -f $(BIN)
	@# Remove cross-platform release outputs from the distribution directory.
	rm -rf $(DIST_DIR)
	@# Remove workspace dependency installs and Turbo cache state.
	rm -rf .turbo node_modules tooling/node_modules semantic/node_modules
	@# Clear the Go build cache for the semantic workspace.
	go -C $(GO_WORKDIR) clean -cache
