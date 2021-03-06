#!/usr/bin/env bash

set -e

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../" >/dev/null 2>&1 && pwd)"

mkdir -p ${project_root}/tools/dev/bin
mkdir -p ${project_root}/tools/dev/db_data

docker-compose -f ${project_root}/tools/dev/docker-compose.yml  --project-directory ${project_root} "$@"