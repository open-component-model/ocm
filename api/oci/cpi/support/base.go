package support

import (
	"fmt"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type artifactBase struct {
	lock      sync.RWMutex
	container NamespaceAccessImpl // this is the underlying container implementation
	state     accessobj.State
}

func newArtifactBase(container NamespaceAccessImpl, state accessobj.State) artifactBase {
	return artifactBase{
		container: container,
		state:     state,
	}
}

func (a *artifactBase) IsReadOnly() bool {
	return a.container.IsReadOnly()
}

func (a *artifactBase) IsIndex() bool {
	d, ok := a.state.GetState().(*artdesc.Artifact)
	return ok && d.IsIndex()
}

func (a *artifactBase) IsManifest() bool {
	d, ok := a.state.GetState().(*artdesc.Artifact)
	return ok && d.IsManifest()
}

func (a *artifactBase) IsValid() bool {
	d, ok := a.state.GetState().(*artdesc.Artifact)
	return ok && d.IsValid()
}

func (a *artifactBase) blob() (cpi.BlobAccess, error) {
	return a.state.GetBlob()
}

func (a *artifactBase) Blob() (cpi.BlobAccess, error) {
	d, ok := a.state.GetState().(artdesc.BlobDescriptorSource)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to artdesc.BlobDescriptorSource", a.state.GetState())
	}
	if !d.IsValid() {
		return nil, errors.ErrUnknown("artifact type")
	}
	blob, err := a.blob()
	if err != nil {
		return nil, err
	}
	return blobaccess.WithMimeType(d.MimeType(), blob), nil
}

func (a *artifactBase) Digest() digest.Digest {
	d := a.state.GetState().(artdesc.BlobDescriptorSource)
	if !d.IsValid() {
		return ""
	}
	blob, err := a.blob()
	if err != nil {
		return ""
	}
	return blob.Digest()
}
