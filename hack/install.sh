#!/bin/bash

# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

set -e

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..
GIT_TREE_STATE="$([ -z "$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)"

if [[ $EFFECTIVE_VERSION == "" ]]; then
  EFFECTIVE_VERSION=$(cat $PROJECT_ROOT/VERSION)
fi

CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) GO111MODULE=on \
  go install \
  -ldflags "-s -w \
            -X github.com/open-component-model/ocm/pkg/version.gitVersion=$EFFECTIVE_VERSION \
            -X github.com/open-component-model/ocm/pkg/version.gitTreeState=$GIT_TREE_STATE \
            -X github.com/open-component-model/ocm/pkg/version.gitCommit=$(git rev-parse --verify HEAD) \
            -X github.com/open-component-model/ocm/pkg/version.buildDate=$(date -u +%FT%T%z)" \
  ${PROJECT_ROOT}/cmds/ocm
