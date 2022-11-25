// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package consts

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const (
	// OCIArtifact describes a generic OCI artifact following the
	//   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md)
	OCIArtifact = resourcetypes.OCI_ARTIFACT
	// OCIImage describes an OCIArtifact containing an image.
	OCIImage = resourcetypes.OCI_IMAGE
	// HelmChart describes a helm chart, either stored as OCI artifact or as tar blob (tar media type).
	HelmChart = resourcetypes.HELM_CHART
	// blob describes any anonymous untyped blob data.
	Blob = resourcetypes.BLOB
	// FileSystem describes a directory structure stored as archive (tar, tgz).
	FileSystem = resourcetypes.FILESYSTEM
	// Executable describes an OS executable.
	Executable = resourcetypes.EXECUTABLE
	// OCMPlugin describes an OS executable OCM plugin.
	OCMPlugin = resourcetypes.OCM_PLUGIN
)
