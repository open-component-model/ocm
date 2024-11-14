#!/bin/bash

set -e

# This script is used to get a bare resource from a CTF file.
# It can be used in case the OCM CLI is not available to extract the resource from a CTF.
# A typical use case for this is the "OCM Inception" in which a CTF containing the CLI needs to be extracted
# to run the CLI to extract the resource.
#
# In this case one can use this script to extract the correct OCM CLI without having to rely on the CLI being
# already available.
#
# By default the script will look for the OCM CLI component with any version (the first encountered will be used)
# and will extract the resource "ocmcli" for amd64/linux as a filepath. This path can then be used to run the CLI,
# but only after allowing to execute it, e.g with `chmod +x <path>`.

COMPONENT=${1:-"ocm.software/ocmcli"}
COMPONENT_VERSION=${2:-""}
RESOURCE=${3:-"ocmcli"}
ARCHITECTURE=${4:-"amd64"}
OS=${5:-"linux"}
MEDIA_TYPE=${6:-"application/octet-stream"}
PATH_TO_CTF=${7:-"./gen/ctf"}

INDEX=$( \
yq -r ".artifacts | filter(.repository == \"component-descriptors/${COMPONENT}\" and (.tag | contains(\"${COMPONENT_VERSION}\")))[0].digest" \
  "${PATH_TO_CTF}"/artifact-index.json | \
  sed 's/:/./g' \
)

if [ -z "${INDEX}" ]; then
  echo "No index found for ${COMPONENT}"
  exit 1
fi

RESOURCE=$( \
yq ".layers | filter(
    (
      .annotations.\"software.ocm.artifact\" |
      from_json |
      .[0]
    ) as \$artifact |
    (
      \$artifact.identity.name == \"$RESOURCE\" and
      \$artifact.identity.architecture == \"$ARCHITECTURE\" and
      \$artifact.identity.os == \"$OS\" and
      .mediaType == \"$MEDIA_TYPE\"
    )
  )[0].digest" "${PATH_TO_CTF}"/blobs/"${INDEX}" | sed 's/:/./g' \
)

if [ -z "${RESOURCE}" ]; then
  echo "No resource found for ${COMPONENT}"
  exit 1
fi

RESOURCE=$PATH_TO_CTF/blobs/$RESOURCE

echo "$RESOURCE"
