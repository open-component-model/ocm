#!/bin/bash
#
# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

GOLANGCI_LINT_CONFIG_FILE=""

for arg in "$@"; do
  case $arg in
    --golangci-lint-config=*)
    GOLANGCI_LINT_CONFIG_FILE="-c ${arg#*=}"
    shift
    ;;
  esac
done

echo "> Check"

echo "Executing golangci-lint"
echo "  golangci-lint run $GOLANGCI_LINT_CONFIG_FILE --timeout 10m $@"
golangci-lint run $GOLANGCI_LINT_CONFIG_FILE --timeout 10m $@

echo "Executing gofmt"
folders=()
for f in $@; do
  folders+=( "$(echo $f | sed 's/\(.*\)\/\.\.\./\1/')" )
done
unformatted_files="$(goimports -l -local=github.com/open-component-model/ocm ${folders[*]})"
if [[ "$unformatted_files" ]]; then
  echo "Unformatted files detected:"
  echo "$unformatted_files"
  exit 1
fi

echo "All checks successful"
