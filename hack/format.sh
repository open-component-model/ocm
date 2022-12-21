#!/bin/bash

# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

set -e

log() {
  msg=${1:-(no message)}

  echo " === ${msg}"
}

pkgprefix="github.com/open-component-model/ocm"

log "Format with gci"
GCIFMT=( -s standard -s blank -s dot -s default -s="prefix(${pkgprefix})" --custom-order )
gci diff --skip-generated "${GCIFMT[@]}"  $@ </dev/null \
  | awk '/^--- / { print $2 }' \
  | xargs -I "{}" \
    gci write --skip-generated  "${GCIFMT[@]}" "{}"
log "Format done"
