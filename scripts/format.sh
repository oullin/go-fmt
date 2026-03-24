#!/usr/bin/env bash
set -euo pipefail

output="${1:-text}"
config="${2:-}"
shift 2 || true

args=("$@")
if [[ ${#args[@]} -eq 0 ]]; then
	args=(.)
fi

cmd=(go run ./cmd/fmt format --format "$output")
if [[ -n "$config" ]]; then
	cmd+=(--config "$config")
fi
cmd+=("${args[@]}")

"${cmd[@]}"
