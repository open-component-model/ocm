#!/bin/bash -e

# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

echo "> Install Go packages/binaries"

curl -sfL "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" | sh -s -- -b $(go env GOPATH)/bin v1.54.2

GO111MODULE=off go get -u github.com/go-bindata/go-bindata/...

go install github.com/go-bindata/go-bindata/v3/go-bindata@v3.1.3
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/daixiang0/gci@v0.7.0
go install github.com/mandelsoft/mdref@master

echo "> Install Registry test binaries"

mkdir -p ${PROJECT_ROOT}/tmp/test/bin
curl -L "https://storage.googleapis.com/gardener-public/test/oci-registry/registry-$(go env GOOS)-$(go env GOARCH)" --output ${PROJECT_ROOT}/tmp/test/bin/registry
chmod +x ${PROJECT_ROOT}/tmp/test/bin/registry


