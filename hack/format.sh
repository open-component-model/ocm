#!/bin/bash
#
# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
gci diff --skip-generated "${GCIFMT[@]}"  $@ \
  | awk '/^--- / { print $2 }' \
  | xargs -I "{}" \
    gci write --skip-generated  "${GCIFMT[@]}" "{}"
