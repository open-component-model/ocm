#!/bin/bash -e
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

echo "> Install Go packages/binaries"

curl -sfL "https://install.goreleaser.com/github.com/golangci/golangci-lint.sh" | sh -s -- -b $(go env GOPATH)/bin v1.32.2

GO111MODULE=off go get golang.org/x/tools/cmd/goimports
GO111MODULE=off go get -u github.com/go-bindata/go-bindata/...

go install golang.org/x/tools/cmd/goimports@latest
go install github.com/daixiang0/gci@v0.7.0

echo "> Install Registry test binaries"

mkdir -p ${PROJECT_ROOT}/tmp/test/bin
curl -L "https://storage.googleapis.com/gardener-public/test/oci-registry/registry-$(go env GOOS)-$(go env GOARCH)" --output ${PROJECT_ROOT}/tmp/test/bin/registry
chmod +x ${PROJECT_ROOT}/tmp/test/bin/registry


