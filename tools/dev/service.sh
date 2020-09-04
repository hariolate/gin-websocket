#!/usr/bin/env bash

set -e

project_root=$($(dirname "${BASH_SOURCE[0]}")/../project_root.sh)

mkdir -p ${project_root}/tools/dev/bin
mkdir -p ${project_root}/tools/dev/db_data

docker-compose -f ${project_root}/tools/dev/docker-compose.yml  --project-directory ${project_root} "$@"