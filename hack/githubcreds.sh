#!/bin/bash
# SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

createAuth() {
  if [ -n "$GITHUB_REPOSITORY_OWNER" -a -n "$GITHUB_TOKEN" ]; then
    ocm_comprepo="ghcr.io/$GITHUB_REPOSITORY_OWNER/ocm"
    ocm_comprepouser=$GITHUB_REPOSITORY_OWNER
    ocm_comprepopassword="$GITHUB_TOKEN"
    comprepourl="${ocm_comprepo#*//}"
    repohost="${comprepourl%%/*}"
    comprepourl="${ocm_comprepo%$comprepourl}${comprepourl%%/*}"
    creds=(--cred :type=OCIRegistry --cred ":hostname=$repohost" --cred "username=$ocm_comprepouser" --cred "password=$ocm_comprepopassword")

    echo "${creds[@]}"
  fi
}

createAuth
