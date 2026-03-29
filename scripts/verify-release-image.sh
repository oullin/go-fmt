#!/usr/bin/env bash
set -euo pipefail

if [ -z "${IMAGE_NAME:-}" ]; then
	printf 'IMAGE_NAME is required\n' >&2
	exit 1
fi

if [ -z "${NEW_TAG:-}" ]; then
	printf 'NEW_TAG is required\n' >&2
	exit 1
fi

version_output="$(docker run --rm "${IMAGE_NAME}:${NEW_TAG}" version)"

if [ "$version_output" != "go-fmt ${NEW_TAG}" ]; then
	printf 'unexpected version output for %s: %s\n' "${IMAGE_NAME}:${NEW_TAG}" "$version_output" >&2
	exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

cat > "${tmpdir}/sample.go" <<'EOF'
package sample

func run() {
	if true {
		println("ok")
	}
	println("next")
}
EOF

set +e
report_output="$(docker run --rm -v "${tmpdir}:/work" -w /work "${IMAGE_NAME}:${NEW_TAG}" check . 2>&1)"
report_status=$?
set -e

printf '%s\n' "$report_output"

if [ "$report_status" -ne 1 ]; then
	printf 'expected check to exit 1, got %s\n' "$report_status" >&2
	exit 1
fi

printf '%s\n' "$report_output" | grep -Fq "  sample.go"
printf '%s\n' "$report_output" | grep -Fq "    [spacing] line 7: missing blank line after if statement"
printf '%s\n' "$report_output" | grep -Fq "  Result: fail. 1 changed, 1 violation(s), 0 error(s)."

if printf '%s\n' "$report_output" | grep -Fq "~ sample.go:7 [spacing]"; then
	printf 'legacy flat renderer detected in %s\n' "${IMAGE_NAME}:${NEW_TAG}" >&2
	exit 1
fi
