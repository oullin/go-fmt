#!/usr/bin/env bash
set -euo pipefail

app="${APP:-fmt}"
go_workdir="${GO_WORKDIR:-semantic}"
cmd="${CMD:-./cmd/fmt}"
build_dir="${BUILD_DIR:-bin}"
bin="${BIN:-${build_dir}/${app}}"
version="${VERSION:-dev}"
cgo_enabled="${CGO_ENABLED:-0}"
host_os="${HOST_OS:-$(go -C "$go_workdir" env GOOS)}"
host_arch="${HOST_ARCH:-$(go -C "$go_workdir" env GOARCH)}"

mkdir -p "$build_dir"

CGO_ENABLED="$cgo_enabled" GOOS="$host_os" GOARCH="$host_arch" \
	go -C "$go_workdir" build -ldflags "-X main.version=$version" -o "../$bin" "$cmd"
chmod +x "$bin"
