#!/usr/bin/env bash

set -e

project_root=$($(dirname "${BASH_SOURCE[0]}")/project_root.sh)

protoc -I="${project_root}"/src/protocol --go_out="${project_root}"/src/protocol "${project_root}/src/protocol/"*.proto
protoc -I="${project_root}"/src/protocol --js_out=import_style=commonjs:"${project_root}"/src/protocol "${project_root}/src/protocol/"*.proto

# workaround "error  'proto' is not defined         no-undef"
for f in "${project_root}"/src/protocol/*.js
do
    echo '/* eslint-disable */' | cat - "${f}" > temp && mv temp "${f}"
done