#!/bin/bash

set -e

GOLANGCI_LINT_CONFIG_FILE=""

if [ "$1" == "--fix" ]; then
  opt="--fix"
  shift
fi


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
echo "  golangci-lint run $GOLANGCI_LINT_CONFIG_FILE $opt --timeout 10m $@"
golangci-lint run $GOLANGCI_LINT_CONFIG_FILE $opt --timeout 10m $@

echo "All checks successful"
