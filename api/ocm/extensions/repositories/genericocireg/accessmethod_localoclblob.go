package genericocireg

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/api/oci"
	"github.com/open-component-model/ocm/api/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/localblob"
	"github.com/open-component-model/ocm/api/utils/refmgmt"
)

type localOCIBlobAccessMethod struct {
	*localBlobAccessMethod
}

var _ accspeccpi.AccessMethodImpl = (*localOCIBlobAccessMethod)(nil)

func newLocalOCIBlobAccessMethod(a *localblob.AccessSpec, ns oci.NamespaceAccess, art oci.ArtifactAccess, ref refmgmt.ExtendedAllocatable) (accspeccpi.AccessMethod, error) {
	m, err := newLocalBlobAccessMethodImpl(a, ns, art, ref)
	return accspeccpi.AccessMethodForImplementation(&localOCIBlobAccessMethod{
		localBlobAccessMethod: m,
	}, err)
}

func (m *localOCIBlobAccessMethod) MimeType() string {
	digest := digest.Digest(m.spec.LocalReference)
	desc := m.artifact.GetDescriptor().GetBlobDescriptor(digest)
	if desc == nil {
		return ""
	}
	return desc.MediaType
}
