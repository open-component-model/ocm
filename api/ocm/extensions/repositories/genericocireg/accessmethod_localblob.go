package genericocireg

import (
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/refmgmt"
)

type localBlobAccessMethod struct {
	lock      sync.Mutex
	err       error
	data      blobaccess.DataAccess
	spec      *localblob.AccessSpec
	namespace oci.NamespaceAccess
	artifact  oci.ArtifactAccess
}

var _ accspeccpi.AccessMethodImpl = (*localBlobAccessMethod)(nil)

func newLocalBlobAccessMethod(a *localblob.AccessSpec, ns oci.NamespaceAccess, art oci.ArtifactAccess, ref refmgmt.ExtendedAllocatable) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newLocalBlobAccessMethodImpl(a, ns, art, ref))
}

func newLocalBlobAccessMethodImpl(a *localblob.AccessSpec, ns oci.NamespaceAccess, art oci.ArtifactAccess, ref refmgmt.ExtendedAllocatable) (*localBlobAccessMethod, error) {
	m := &localBlobAccessMethod{
		spec:      a,
		namespace: ns,
		artifact:  art,
	}
	ref.BeforeCleanup(refmgmt.CleanupHandlerFunc(m.cache))
	return m, nil
}

func (m *localBlobAccessMethod) cache() {
	if m.artifact != nil {
		_, m.err = m.getBlob()
	}
}

func (_ *localBlobAccessMethod) IsLocal() bool {
	return true
}

func (m *localBlobAccessMethod) GetKind() string {
	return m.spec.GetKind()
}

func (m *localBlobAccessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *localBlobAccessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.artifact = nil
	m.namespace = nil
	if m.data != nil {
		tmp := m.data
		m.data = nil
		return tmp.Close()
	}
	return nil
}

func (m *localBlobAccessMethod) getBlob() (blobaccess.DataAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.data != nil {
		return m.data, nil
	}
	if artdesc.IsOCIMediaType(m.spec.MediaType) {
		// may be we should always store the blob, additionally to the
		// exploded form to make things easier.

		if m.spec.LocalReference == "" {
			// TODO: synthesize the artifact blob
			return nil, errors.ErrNotImplemented("artifact blob synthesis")
		}
	}
	_, data, err := m.namespace.GetBlobData(digest.Digest(m.spec.LocalReference))
	if err != nil {
		return nil, err
	}
	m.data = data
	return m.data, err
}

func (m *localBlobAccessMethod) Reader() (io.ReadCloser, error) {
	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}

func (m *localBlobAccessMethod) Get() ([]byte, error) {
	return blobaccess.BlobData(m.getBlob())
}

func (m *localBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
