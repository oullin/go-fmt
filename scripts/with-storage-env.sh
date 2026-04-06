#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

should_cleanup_turbo_logs=0
if [[ "${1:-}" == "pnpm" && "${2:-}" == "turbo" ]] || [[ "${1:-}" == "turbo" ]]; then
	should_cleanup_turbo_logs=1
fi

ensure_storage_layout
set +e
"$@"
status=$?
set -e
if [[ "$should_cleanup_turbo_logs" -eq 1 ]]; then
	cleanup_package_turbo_logs
	assert_no_package_turbo_dirs
fi
assert_no_legacy_artifacts
exit "$status"
