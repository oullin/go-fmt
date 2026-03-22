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
HOST_ARCH := $(shell case "$$(uname -m)" in \
	x86_64|amd64) echo amd64 ;; \
	arm64|aarch64) echo arm64 ;; \
	*) echo "$$(uname -m)" ;; \
esac)
RELEASE_PLATFORMS := darwin/arm64 linux/amd64

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
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) go build -ldflags "-X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o $(BIN) $(CMD)

release:
	mkdir -p $(DIST_DIR)
	@for platform in $(RELEASE_PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		case $${GOOS} in \
			darwin) os_label="macOS"; os_slug="macos" ;; \
			linux)  os_label="Linux"; os_slug="linux" ;; \
			*)      os_label=$${GOOS}; os_slug=$${GOOS} ;; \
		esac; \
		case $${GOARCH} in \
			amd64) arch_label="x86_64"; arch_slug="x86_64" ;; \
			arm64) arch_label="Apple Silicon"; arch_slug="apple-silicon" ;; \
			*)     arch_label=$${GOARCH}; arch_slug=$${GOARCH} ;; \
		esac; \
		output="$(DIST_DIR)/$(APP)-$${os_slug}-$${arch_slug}"; \
		echo "Building $${os_label} $${arch_label} ($${GOOS}/$${GOARCH})..."; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$${GOOS} GOARCH=$${GOARCH} go build -ldflags "-X main.version=$(VERSION)" -o "$${output}" $(CMD); \
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
