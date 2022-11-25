// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"github.com/open-component-model/ocm/pkg/common/accessobj"
)

const (
	DirMode  = accessobj.DirMode
	FileMode = accessobj.FileMode
)

var ModTime = accessobj.ModTime

const (
	// The artifact descriptor name for artifact format.
	ArtifactSetDescriptorFileName = "artifact-descriptor.json"
	// BlobsDirectoryName is the name of the directory holding the artifact archives.
	BlobsDirectoryName = "blobs"
	// ArtifactIndexFileName is the artifact index descriptor name for CommanTransportFormat.
	ArtifactIndexFileName = "artifact-index.json"
)
