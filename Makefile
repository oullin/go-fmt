SHELL := /bin/zsh

APP := fmt
CMD := ./cmd/fmt
BUILD_DIR := builds
BIN := $(BUILD_DIR)/$(APP)
ARGS ?= .
CONFIG ?=
OUTPUT ?= text
GOIMPORTS := golang.org/x/tools/cmd/goimports@latest
DIST_DIR := dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

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
	go build -ldflags "-X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o $(BIN) $(CMD)

release:
	mkdir -p $(DIST_DIR)
	@for platform in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		output="$(DIST_DIR)/$(APP)-$${GOOS}-$${GOARCH}"; \
		echo "Building $${GOOS}/$${GOARCH}..."; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} go build -ldflags "-X main.version=$(VERSION)" -o "$${output}" $(CMD); \
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
