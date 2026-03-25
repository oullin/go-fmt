#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

build_dir="${BUILD_DIR:-bin}"
bin="${BIN:-${build_dir}/${APP}}"
host_os="${HOST_OS:-$(go -C "$GO_WORKDIR" env GOOS)}"
host_arch="${HOST_ARCH:-$(go -C "$GO_WORKDIR" env GOARCH)}"

mkdir -p "$build_dir"

CGO_ENABLED="$CGO_ENABLED" GOOS="$host_os" GOARCH="$host_arch" \
	go -C "$GO_WORKDIR" build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "../$bin" "$CMD"
chmod +x "$bin"

