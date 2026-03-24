#!/usr/bin/env bash
set -euo pipefail

if [[ -e go-fmt.yml ]]; then
	printf 'config already exists at ./go-fmt.yml\n'
	exit 0
fi

cp go-fmt.yml.example go-fmt.yml
printf 'config ready at ./go-fmt.yml\n'
