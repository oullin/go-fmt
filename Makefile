SHELL := /bin/bash

APP := fmt
CMD := ./cmd/fmt
BUILD_DIR := bin
BIN := $(BUILD_DIR)/$(APP)
ARGS ?= .
CONFIG ?=
OUTPUT ?= text
GOIMPORTS := golang.org/x/tools/cmd/goimports@latest
DIST_DIR := dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
CGO_ENABLED ?= 0
HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH := $(shell arch="$$(uname -m)"; if [ "$$arch" = "x86_64" ] || [ "$$arch" = "amd64" ]; then echo amd64; elif [ "$$arch" = "arm64" ] || [ "$$arch" = "aarch64" ]; then echo arm64; else echo "$$arch"; fi)
RELEASE_PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: help run check format check-json check-agent config build release test test-race test-short vet lint install install-tools clean

help:
	@echo "go-fmt developer targets"
	@echo ""
	@echo "Usage:"
	@echo "  make run"
	@echo "  make check"
	@echo "  make format"
	@echo "  make check ARGS=./rules/spacing/spacing.go"
	@echo "  make format ARGS=./rules/spacing/spacing.go"
	@echo "  make check-json"
	@echo "  make check-agent"
	@echo "  make config"
	@echo "  make build"
	@echo "  make release"
	@echo "  make test"
	@echo ""
	@echo "Variables:"
	@echo "  ARGS=$(ARGS)"
	@echo "  CONFIG=$(CONFIG)"
	@echo "  OUTPUT=$(OUTPUT)"

run:
	go run $(CMD) --help

check:
	go run $(CMD) check $(if $(CONFIG),--config $(CONFIG),) --format $(OUTPUT) $(ARGS)

format:
	go run $(CMD) format $(if $(CONFIG),--config $(CONFIG),) --format $(OUTPUT) $(ARGS)

check-json:
	go run $(CMD) check $(if $(CONFIG),--config $(CONFIG),) --format json $(ARGS)

check-agent:
	go run $(CMD) check $(if $(CONFIG),--config $(CONFIG),) --format agent $(ARGS)

config:
	cp -n go-fmt.yml.example go-fmt.yml || true
	@echo "config ready at ./go-fmt.yml"

build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) go build -ldflags "-X main.version=$(VERSION)" -o $(BIN) $(CMD)
	chmod +x $(BIN)

release:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	@for platform in $(RELEASE_PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		case $${GOOS} in \
			darwin) os_label="macOS" ;; \
			linux)  os_label="Linux" ;; \
			*)      os_label=$${GOOS} ;; \
		esac; \
		case $${GOARCH} in \
			amd64) arch_label="x86_64" ;; \
			arm64) arch_label="arm64" ;; \
			*)     arch_label=$${GOARCH} ;; \
		esac; \
		if [ "$${GOOS}" = "darwin" ] && [ "$${GOARCH}" = "arm64" ]; then arch_label="Apple Silicon"; fi; \
		output="$(DIST_DIR)/$(APP)-$${GOOS}-$${GOARCH}"; \
		echo "Building $${os_label} $${arch_label} ($${GOOS}/$${GOARCH})..."; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$${GOOS} GOARCH=$${GOARCH} go build -ldflags "-X main.version=$(VERSION)" -o "$${output}" $(CMD); \
		chmod +x "$${output}"; \
	done

test:
	go test ./... -v

test-race:
	go test ./... -race -v

test-short:
	go test ./... -short

vet:
	go vet ./...

lint:
	gofmt -w .
	go test ./...
	go vet ./...

install:
	go install $(CMD)

install-tools:
	go install $(GOIMPORTS)

clean:
	rm -f $(BIN)
	rm -rf $(DIST_DIR)
	go clean -cache
