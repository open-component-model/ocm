// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package consts

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const (
	// OCIArtefact describes a generic OCI artefact following the
	//   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md)
	OCIArtefact = resourcetypes.OCI_ARTEFACT
	// OCIImage describes an OCIArtefact containing an image.
	OCIImage = resourcetypes.OCI_IMAGE
	// HelmChart describes a helm chart, either stored as OCI artefact or as tar blob (tar media type).
	HelmChart = resourcetypes.HELM_CHART
	// blob describes any anonymous untyped blob data.
	Blob = resourcetypes.BLOB
	// FileSystem describes a directory structure stored as archive (tar, tgz).
	FileSystem = resourcetypes.FILESYSTEM
)
