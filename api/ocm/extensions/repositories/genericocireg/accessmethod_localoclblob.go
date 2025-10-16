package genericocireg

import (
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/utils/refmgmt"
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
