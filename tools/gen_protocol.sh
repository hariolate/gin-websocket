#!/usr/bin/env bash

set -e

project_root=$($(dirname "${BASH_SOURCE[0]}")/project_root.sh)

protoc -I="${project_root}/src/protocol" --go_out="${project_root}/src/service/protocol" "${project_root}/src/protocol/"*.proto