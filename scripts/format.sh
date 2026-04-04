#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
oxfmt_bin="${OXFMT_BIN:-tooling/node_modules/.bin/oxfmt}"

declare -a args=("$@")
declare -a semantic_args=()
declare -a selected_args=()

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

is_repo_root_selector() {
	local arg="$1"

	case "$arg" in
		.|"$repo_root"|"$GO_WORKDIR"|"$repo_root/$GO_WORKDIR")
			return 0
			;;
		*)
			return 1
			;;
	esac
}

use_diff_selection=true

for raw_arg in "${args[@]}"; do
	arg="$(normalize_arg "$raw_arg")"
	if ! is_repo_root_selector "$arg"; then
		use_diff_selection=false
		break
	fi
done

if [[ "$use_diff_selection" == true ]] && git -C "$repo_root" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
	while IFS= read -r -d '' path; do
		selected_args+=("$path")
	done < <(
		git -C "$repo_root" diff --name-only --diff-filter=ACMR -z HEAD --
		git -C "$repo_root" ls-files --others --exclude-standard -z
	)
else
	while IFS= read -r -d '' path; do
		selected_args+=("$path")
	done < <(git -C "$repo_root" ls-files --cached --others --exclude-standard -z -- "${args[@]}")
fi

for raw_arg in "${selected_args[@]}"; do
	arg="$(normalize_arg "$raw_arg")"
	if semantic_arg="$(to_semantic_arg "$arg")"; then
		semantic_args+=("$semantic_arg")
	fi
done

if [[ ${#semantic_args[@]} -gt 0 ]]; then
	if [[ "$use_diff_selection" == true ]]; then
		go -C "$GO_WORKDIR" run "$CMD" format --cwd . --git-diff
	else
		go -C "$GO_WORKDIR" run "$CMD" format --cwd . "${semantic_args[@]}"
	fi
fi

if [[ ${#selected_args[@]} -gt 0 ]]; then
	printf '%s\0' "${selected_args[@]}" | xargs -0 "$oxfmt_bin" --write --no-error-on-unmatched-pattern
fi
