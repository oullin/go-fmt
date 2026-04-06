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

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
source_compose="${repo_root}/examples/consumer/go-fmt.compose.yaml"
image_ref="${IMAGE_NAME}:${NEW_TAG}"
cache_volume="go-fmt-cache"
cache_volume_existed=0

if docker volume inspect "${cache_volume}" >/dev/null 2>&1; then
	cache_volume_existed=1
fi

tmpdir="$(mktemp -d)"
cleanup() {
	rm -rf "${tmpdir}"

	if [ "${cache_volume_existed}" -eq 0 ]; then
		docker volume rm "${cache_volume}" >/dev/null 2>&1 || true
	fi
}
trap cleanup EXIT

make_project() {
	local project_dir="$1"

	mkdir -p "${project_dir}/core" "${project_dir}/demo/api"

	cat > "${project_dir}/go.mod" <<'EOF'
module example.com/consumer

go 1.25
EOF

	cat > "${project_dir}/core/sample.go" <<'EOF'
package core

func Run() {
	if true {
		println("ok")
	}
	println("next")
}
EOF

	cat > "${project_dir}/demo/api/sample.go" <<'EOF'
package api

func Run() {
	if true {
		println("ok")
	}
	println("next")
}
EOF
}

make_compose() {
	local destination="$1"

	sed "s|image: ghcr.io/oullin/go-fmt:latest|image: ${image_ref}|" "${source_compose}" > "${destination}"
}

assert_contains() {
	local haystack="$1"
	local needle="$2"
	local label="$3"

	if ! grep -Fq -- "${needle}" <<<"${haystack}"; then
		printf 'expected %s to contain %s\n' "${label}" "${needle}" >&2
		exit 1
	fi
}

expected_sample() {
	local package_name="$1"
	local formatted="$2"

	if [ "${formatted}" = "formatted" ]; then
		cat <<EOF
package ${package_name}

func Run() {
	if true {
		println("ok")
	}

	println("next")
}
EOF
		return
	fi

	cat <<EOF
package ${package_name}

func Run() {
	if true {
		println("ok")
	}
	println("next")
}
EOF
}

assert_file_equals() {
	local path="$1"
	local expected="$2"
	local label="$3"
	local actual

	actual="$(cat "${path}")"

	if [ "${actual}" != "${expected}" ]; then
		printf 'unexpected content for %s (%s)\n' "${label}" "${path}" >&2
		printf '--- actual ---\n%s\n' "${actual}" >&2
		printf '--- expected ---\n%s\n' "${expected}" >&2
		exit 1
	fi
}

run_expect_status() {
	local expected_status="$1"
	local output
	local status
	shift

	set +e
	output="$("$@" 2>&1)"
	status=$?
	set -e

	if [ "${status}" -ne "${expected_status}" ]; then
		printf '%s\n' "${output}"
		printf 'expected exit %s, got %s\n' "${expected_status}" "${status}" >&2
		exit 1
	fi

	printf '%s\n' "${output}"
}

copy_project="${tmpdir}/copy-project"
shared_project="${tmpdir}/shared-project"
outside_dir="${tmpdir}/outside"

make_project "${copy_project}"
make_project "${shared_project}"
mkdir -p "${outside_dir}"
printf 'outside\n' > "${outside_dir}/sample.txt"

copy_compose="${copy_project}/go-fmt.compose.yaml"
shared_compose="${tmpdir}/shared/go-fmt.compose.yaml"
mkdir -p "$(dirname "${shared_compose}")"

make_compose "${copy_compose}"
make_compose "${shared_compose}"

copy_config="$(
	cd "${copy_project}" &&
	docker compose -f go-fmt.compose.yaml config
)"
printf '%s\n' "${copy_config}"
assert_contains "${copy_config}" "source: ${copy_project}" "copied compose config"
assert_contains "${copy_config}" "HOST_PROJECT_PATH: ${copy_project}" "copied compose config"
assert_contains "${copy_config}" "target: /cache" "copied compose config"
assert_contains "${copy_config}" "GOCACHE: /cache/go-build" "copied compose config"
assert_contains "${copy_config}" "name: go-fmt-cache" "copied compose config"

copy_help="$(
	cd "${copy_project}" &&
	run_expect_status 0 docker compose -f go-fmt.compose.yaml run --rm go-fmt help
)"
assert_contains "${copy_help}" "go-fmt check [--host-path /absolute/host/path] [paths...]" "copied compose help output"

copy_check="$(
	cd "${copy_project}" &&
	run_expect_status 1 docker compose -f go-fmt.compose.yaml run --rm go-fmt check ./core ./demo/api
)"
assert_contains "${copy_check}" "Result: fail." "copied compose check output"
assert_contains "${copy_check}" "core/sample.go" "copied compose check output"
assert_contains "${copy_check}" "demo/api/sample.go" "copied compose check output"

copy_format="$(
	cd "${copy_project}" &&
	run_expect_status 0 docker compose -f go-fmt.compose.yaml run --rm go-fmt format ./core ./demo/api
)"
assert_contains "${copy_format}" "Result: fixed." "copied compose format output"

assert_file_equals "${copy_project}/core/sample.go" "$(expected_sample core formatted)" "copied compose formatted core file"
assert_file_equals "${copy_project}/demo/api/sample.go" "$(expected_sample api formatted)" "copied compose formatted api file"

shared_config="$(
	GO_FMT_PROJECT_DIR="${shared_project}" docker compose -f "${shared_compose}" config
)"
printf '%s\n' "${shared_config}"
assert_contains "${shared_config}" "source: ${shared_project}" "shared compose config"
assert_contains "${shared_config}" "HOST_PROJECT_PATH: ${shared_project}" "shared compose config"
assert_contains "${shared_config}" "target: /cache" "shared compose config"
assert_contains "${shared_config}" "GOMODCACHE: /cache/gopath/pkg/mod" "shared compose config"
assert_contains "${shared_config}" "name: go-fmt-cache" "shared compose config"

shared_help="$(
	GO_FMT_PROJECT_DIR="${shared_project}" run_expect_status 0 docker compose -f "${shared_compose}" run --rm go-fmt help
)"
assert_contains "${shared_help}" "go-fmt format [--host-path /absolute/host/path] [paths...]" "shared compose help output"

shared_host_path_format="$(
	GO_FMT_PROJECT_DIR="${shared_project}" run_expect_status 0 \
		docker compose -f "${shared_compose}" run --rm go-fmt format --host-path "${shared_project}/core"
)"
assert_contains "${shared_host_path_format}" "Result: fixed." "shared compose host-path format output"
assert_file_equals "${shared_project}/core/sample.go" "$(expected_sample core formatted)" "shared compose host-path formatted core file"
assert_file_equals "${shared_project}/demo/api/sample.go" "$(expected_sample api original)" "shared compose untouched api file"

shared_invalid_host_path="$(
	GO_FMT_PROJECT_DIR="${shared_project}" run_expect_status 1 \
		docker compose -f "${shared_compose}" run --rm go-fmt check --host-path "${outside_dir}"
)"
assert_contains "${shared_invalid_host_path}" "--host-path must be within HOST_PROJECT_PATH" "shared compose invalid host-path output"

shared_check="$(
	GO_FMT_PROJECT_DIR="${shared_project}" run_expect_status 1 \
		docker compose -f "${shared_compose}" run --rm go-fmt check ./core ./demo/api
)"
assert_contains "${shared_check}" "Result: fail." "shared compose check output"
assert_contains "${shared_check}" "demo/api/sample.go" "shared compose check output"

shared_format="$(
	GO_FMT_PROJECT_DIR="${shared_project}" run_expect_status 0 \
		docker compose -f "${shared_compose}" run --rm go-fmt format ./core ./demo/api
)"
assert_contains "${shared_format}" "Result: fixed." "shared compose format output"
assert_file_equals "${shared_project}/demo/api/sample.go" "$(expected_sample api formatted)" "shared compose formatted api file"
