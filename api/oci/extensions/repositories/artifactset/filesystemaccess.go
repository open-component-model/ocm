package artifactset

import (
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/cpi/support"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type FileSystemBlobAccess struct {
	*accessobj.FileSystemBlobAccess
}

func NewFileSystemBlobAccess(access *accessobj.AccessObject) *FileSystemBlobAccess {
	return &FileSystemBlobAccess{accessobj.NewFileSystemBlobAccess(access)}
}

func (i *FileSystemBlobAccess) GetArtifact(access support.NamespaceAccessImpl, digest digest.Digest) (acc cpi.ArtifactAccess, err error) {
	v, err := access.View()
	if err != nil {
		return nil, err
	}
	defer v.Close()
	_, data, err := i.GetBlobData(digest)
	if err == nil {
		blob := blobaccess.ForDataAccess("", -1, "", data)
		acc, err = support.NewArtifactForBlob(access, blob)
	}
	return acc, err
}

func (i *FileSystemBlobAccess) AddArtifactBlob(artifact cpi.Artifact) (cpi.BlobAccess, error) {
	blob, err := artifact.Blob()
	if err != nil {
		return nil, err
	}

	err = i.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return blob, nil
}
