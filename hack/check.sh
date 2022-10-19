#!/bin/bash

# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
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

echo "All checks successful"
