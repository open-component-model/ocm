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
	// The artefact descriptor name for artefact format.
	ArtefactSetDescriptorFileName = "artefact-descriptor.json"
	// BlobsDirectoryName is the name of the directory holding the artefact archives.
	BlobsDirectoryName = "blobs"
	// ArtefactIndexFileName is the artefact index descriptor name for CommanTransportFormat.
	ArtefactIndexFileName = "artefact-index.json"
)
