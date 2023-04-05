#!/bin/bash
# SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

config() {
  cat <<EOF
type: generic.config.ocm.software/v1
configurations:
  - type: credentials.config.ocm.software
    consumers:
      - identity:
          type: OCIRegistry
        credentials:
          - type: Credentials
            properties:
              hostname: $repohost
              username: $ocm_comprepouser
              password: $ocm_comprepopassword
EOF
}

createAuth() {
  if [ -n "$GITHUB_REPOSITORY_OWNER" -a -n "$GITHUB_TOKEN" ]; then
    ocm_comprepo="ghcr.io/$GITHUB_REPOSITORY_OWNER/ocm"
    ocm_comprepouser=$GITHUB_REPOSITORY_OWNER
    ocm_comprepopassword="$GITHUB_TOKEN"
    comprepourl="${ocm_comprepo#*//}"
    repohost="${comprepourl%%/*}"
    comprepourl="${ocm_comprepo%$comprepourl}${comprepourl%%/*}"
    #creds=(--cred :type=OCIRegistry --cred ":hostname=$repohost" --cred "username=$ocm_comprepouser" --cred "password=$ocm_comprepopassword")
    mkdir -p gen
    config > gen/.ocmconfig
    creds=( --config gen/.ocmconfig )
    echo "${creds[@]}"
  fi
}

createAuth
