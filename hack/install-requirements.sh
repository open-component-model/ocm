#!/bin/bash
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

curl -sfL "https://install.goreleaser.com/github.com/golangci/golangci-lint.sh" | sh -s -- -b $(go env GOPATH)/bin v1.32.2

GO111MODULE=off go get golang.org/x/tools/cmd/goimports

echo "> Install Registry test binaries"

mkdir -p ${PROJECT_ROOT}/tmp/test/bin
curl -L "https://storage.googleapis.com/gardener-public/test/oci-registry/registry-$(go env GOOS)-$(go env GOARCH)" --output ${PROJECT_ROOT}/tmp/test/bin/registry
chmod +x ${PROJECT_ROOT}/tmp/test/bin/registry

platform=$(uname -s)
if [[ ${platform} == *"Darwin"* ]]; then
  cat <<EOM
You are running in a MAC OS environment!
Please make sure you have installed the following requirements:
- GNU Core Utils
- GNU Tar
- GNU Sed
Brew command:
$ brew install coreutils gnu-sed gnu-tar grep jq
Please allow them to be used without their "g" prefix:
$ export PATH=/usr/local/opt/coreutils/libexec/gnubin:\$PATH
$ export PATH=/usr/local/opt/gnu-sed/libexec/gnubin:\$PATH
$ export PATH=/usr/local/opt/gnu-tar/libexec/gnubin:\$PATH
$ export PATH=/usr/local/opt/grep/libexec/gnubin:\$PATH
EOM
fi

