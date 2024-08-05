#!/bin/bash

set -e

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

if [[ $EFFECTIVE_VERSION == "" ]]; then
  EFFECTIVE_VERSION=$(cat $PROJECT_ROOT/VERSION)
fi

mkdir -p dist

build_matrix=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for i in "${build_matrix[@]}"; do
  IFS='/' read os arch <<< "${i}"

  echo "Build $os/$arch"
  bin_path="dist/ocm-$os-$arch"

  CGO_ENABLED=0 GOOS=$os GOARCH=$arch GO111MODULE=on \
  go build -o $bin_path \
  -ldflags "-s -w \
            -X ocm.software/ocm/api/version.gitVersion=$EFFECTIVE_VERSION \
            -X ocm.software/ocm/api/version.gitTreeState=$([ -z "$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty) \
            -X ocm.software/ocm/api/version.gitCommit=$(git rev-parse --verify HEAD) \
            -X ocm.software/ocm/api/version.buildDate=$(date -u +%FT%T%z)" \
  ${PROJECT_ROOT}/cmds/ocm

  # create zipped file
  (cd dist; tar -cvzf "ocm-$os-$arch.tgz" "ocm-$os-$arch")
done
