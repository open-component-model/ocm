package format

import (
	"github.com/open-component-model/ocm/api/oci/extensions/repositories/artifactset"
	"github.com/open-component-model/ocm/api/utils/accessobj"
)

const (
	DirMode  = accessobj.DirMode
	FileMode = accessobj.FileMode
)

var ModTime = accessobj.ModTime

const (
	// BlobsDirectoryName is the name of the directory holding the artifact archives.
	BlobsDirectoryName = artifactset.BlobsDirectoryName
	// ArtifactIndexFileName is the artifact index descriptor name for CommanTransportFormat.
	ArtifactIndexFileName = "artifact-index.json"
)
