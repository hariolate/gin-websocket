#!/usr/bin/env bash

project_root=$($(dirname "${BASH_SOURCE[0]}")/project_root.sh)

mkdir -p "${project_root}"/bin
mkdir -p "${project_root}"/static

"${project_root}"/tools/gen_protocol.sh

go build -v -o "${project_root}"/bin/app-$(uname) "${project_root}"/src/cmd/app

npm --prefix ${project_root} install
npm --prefix ${project_root} run build