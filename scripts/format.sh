#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
oxfmt_bin="${OXFMT_BIN:-tooling/node_modules/.bin/oxfmt}"

declare -a args=("$@")
declare -a semantic_args=()

if [[ ${#args[@]} -eq 0 ]]; then
	args=(.)
fi

normalize_arg() {
	local arg="$1"

	case "$arg" in
		./)
			printf '.\n'
			;;
		./*)
			printf '%s\n' "${arg#./}"
			;;
		*)
			printf '%s\n' "$arg"
			;;
	esac
}

to_semantic_arg() {
	local arg="$1"

	case "$arg" in
		"$repo_root")
			printf '.\n'
			;;
		"$repo_root/$GO_WORKDIR")
			printf '.\n'
			;;
		"$repo_root/$GO_WORKDIR/"*)
			printf '%s\n' "${arg#"$repo_root/$GO_WORKDIR"/}"
			;;
		.)
			printf '.\n'
			;;
		"$GO_WORKDIR")
			printf '.\n'
			;;
		"$GO_WORKDIR/"*)
			printf '%s\n' "${arg#"$GO_WORKDIR"/}"
			;;
		*)
			return 1
			;;
	esac
}

for raw_arg in "${args[@]}"; do
	arg="$(normalize_arg "$raw_arg")"
	if semantic_arg="$(to_semantic_arg "$arg")"; then
		semantic_args+=("$semantic_arg")
	fi
done

if [[ ${#semantic_args[@]} -gt 0 ]]; then
	go -C "$GO_WORKDIR" run "$CMD" format --cwd . "${semantic_args[@]}"
fi

git ls-files --cached --others --exclude-standard -z -- "${args[@]}" \
	| xargs -0 "$oxfmt_bin" --write --no-error-on-unmatched-pattern
