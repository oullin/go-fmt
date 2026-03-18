SHELL := /bin/zsh

APP := fmt
CMD := ./cmd/fmt
BUILD_DIR := builds
BIN := $(BUILD_DIR)/$(APP)
ARGS ?= .
CONFIG ?=
OUTPUT ?= text
GOIMPORTS := golang.org/x/tools/cmd/goimports@latest

.PHONY: help run check format check-json check-agent config build test test-race test-short vet lint install install-tools clean

help:
	@echo "go-cs-fixer developer targets"
	@echo ""
	@echo "Usage:"
	@echo "  make run"
	@echo "  make check"
	@echo "  make format"
	@echo "  make check ARGS=./internal/rules/spacing/spacing.go"
	@echo "  make format ARGS=./internal/rules/spacing/spacing.go"
	@echo "  make check-json"
	@echo "  make check-agent"
	@echo "  make config"
	@echo "  make build"
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
	cp -n go-cs-fixer.yml.example go-cs-fixer.yml || true
	@echo "config ready at ./go-cs-fixer.yml"

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o $(BIN) $(CMD)

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
	go clean -cache
