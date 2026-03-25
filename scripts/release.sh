#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

dist_dir="${DIST_DIR:-dist}"
release_platforms="${RELEASE_PLATFORMS:-darwin/amd64 darwin/arm64 linux/amd64 linux/arm64}"

os_label() {
	case "$1" in
		darwin) printf 'macOS' ;;
		linux) printf 'Linux' ;;
		*) printf '%s' "$1" ;;
	esac
}

arch_label() {
	case "$1/$2" in
		darwin/arm64) printf 'Apple Silicon' ;;
		*/amd64) printf 'x86_64' ;;
		*/arm64) printf 'arm64' ;;
		*) printf '%s' "$2" ;;
	esac
}

rm -rf "$dist_dir"
mkdir -p "$dist_dir"

for platform in $release_platforms; do
	if [[ "$platform" != */* ]]; then
		printf 'invalid release platform: %s\n' "$platform" >&2
		exit 1
	fi

	goos="${platform%/*}"
	goarch="${platform#*/}"
	output="${dist_dir}/${APP}-${goos}-${goarch}"

	printf 'Building %s %s (%s/%s)...\n' "$(os_label "$goos")" "$(arch_label "$goos" "$goarch")" "$goos" "$goarch"

	CGO_ENABLED="$CGO_ENABLED" GOOS="$goos" GOARCH="$goarch" \
		go -C "$GO_WORKDIR" build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "../$output" "$CMD"

	chmod +x "$output"
done
