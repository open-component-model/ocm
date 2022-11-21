// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package resourcetypes

const (
	// OCI_ARTEFACT describes a generic OCI artefact following the
	//   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md).
	OCI_ARTEFACT = "ociArtefact"
	// OCI_IMAGE describes an OCIArtefact containing an image.
	OCI_IMAGE = "ociImage"
	// HELM_CHART describes a helm chart, either stored as OCI artefact or as tar
	// blob (tar media type).
	HELM_CHART = "helmChart"
	// BLOB describes any anonymous untyped blob data.
	BLOB = "blob"
	// FILESYSTEM describes a directory structure stored as archive (tar, tgz).
	FILESYSTEM = "filesystem"
	// EXECUTABLE describes an OS executable.
	EXECUTABLE = "executable"
	// OCM_PLUGIN describes an OS executable OCM plugin.
	OCM_PLUGIN = "ocmPlugin"
)
