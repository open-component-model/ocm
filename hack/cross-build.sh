#!/bin/bash
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

if [[ $EFFECTIVE_VERSION == "" ]]; then
  EFFECTIVE_VERSION=$(cat $PROJECT_ROOT/VERSION)
fi

mkdir -p dist

build_matrix=("linux,amd64" "darwin,amd64" "darwin,arm64" "windows,amd64")

for i in "${build_matrix[@]}"; do
  IFS=',' read os arch <<< "${i}"

  echo "Build $os/$arch"
  bin_path="dist/ocm-$os-$arch"

  CGO_ENABLED=0 GOOS=$os GOARCH=$arch GO111MODULE=on \
  go build -o $bin_path \
  -ldflags "-s -w \
            -X github.com/open-component-model/ocm/pkg/version.gitVersion=$EFFECTIVE_VERSION \
            -X github.com/open-component-model/ocm/pkg/version.gitTreeState=$([ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty) \
            -X github.com/open-component-model/ocm/pkg/version.gitCommit=$(git rev-parse --verify HEAD) \
            -X github.com/open-component-model/ocm/pkg/version.buildDate=$(date --rfc-3339=seconds | sed 's/ /T/')" \
  ${PROJECT_ROOT}/cmds/ocm

  # create zipped file
  gzip -f -k "$bin_path"
done
