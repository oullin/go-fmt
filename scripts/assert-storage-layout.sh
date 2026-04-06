#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

ensure_storage_layout
cleanup_package_turbo_logs
assert_no_package_turbo_dirs
