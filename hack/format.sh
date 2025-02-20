#!/bin/bash

set -e

log() {
  msg=${1:-(no message)}

  echo " === ${msg}"
}

pkgprefix="ocm.software/ocm"

log "Format with gci" # gci write --custom-order --skip-generated -s standard -s blank -s dot -s default -s "Prefix(ocm.software/ocm)"
GCIFMT=( -s standard -s blank -s dot -s default -s="prefix(${pkgprefix})" --custom-order )
go tool gci diff --skip-generated "${GCIFMT[@]}" "${@%/...}" </dev/null \
  | awk '/^--- / { print $2 }' \
  | xargs -I "{}" \
    go tool gci write --skip-generated  "${GCIFMT[@]}" "{}"
log "Format done"

log "Format with gofumpt"
# Specify the pattern or criteria to identify generated files
GENERATED_FILES_PATTERN="*_generated*.go"

# Find all Go files excluding the generated files
directories=( "$@" )
files=()

# Loop through each directory
for dir in "${directories[@]}"; do
  # Search for files in the directory that do not contain the GENERATED_FILES_PATTERN
  files+=( $(find "${dir%/...}" -type f -name "*.go" ! -name "$GENERATED_FILES_PATTERN" ! -name bindata.go) )
done

# Format the files using gofmt with xargs
printf '%s\0' "${files[@]}" | xargs -0 go tool gofumpt -d -e -l -w

log "Format done"
