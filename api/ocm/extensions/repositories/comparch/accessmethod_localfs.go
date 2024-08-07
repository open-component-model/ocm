package comparch

import (
	"io"
	"sync"

	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/refmgmt"
)

////////////////////////////////////////////////////////////////////////////////

type localFilesystemBlobAccessMethod struct {
	sync.Mutex
	closed     bool
	spec       *localblob.AccessSpec
	base       repocpi.ComponentVersionAccessImpl
	err        error
	blobAccess blobaccess.BlobAccess
}

var _ accspeccpi.AccessMethodImpl = (*localFilesystemBlobAccessMethod)(nil)

func newLocalFilesystemBlobAccessMethod(a *localblob.AccessSpec, base repocpi.ComponentVersionAccessImpl, ref refmgmt.ExtendedAllocatable) (accspeccpi.AccessMethod, error) {
	m := &localFilesystemBlobAccessMethod{
		spec: a,
		base: base,
	}
	ref.BeforeCleanup(refmgmt.CleanupHandlerFunc(m.Cache))
	return accspeccpi.AccessMethodForImplementation(m, nil)
}

func (m *localFilesystemBlobAccessMethod) Cache() {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return
	}

	blob, err := m.getBlob()
	if err == nil {
		blob, err = blobaccess.ForCachedBlobAccess(blob, vfsattr.Get(m.base.GetContext()))
	}
	m.blobAccess.Close()
	m.blobAccess = blob
	m.err = err
}

func (_ *localFilesystemBlobAccessMethod) IsLocal() bool {
	return true
}

func (m *localFilesystemBlobAccessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *localFilesystemBlobAccessMethod) GetKind() string {
	return localblob.Type
}

func (m *localFilesystemBlobAccessMethod) Reader() (io.ReadCloser, error) {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return nil, accessio.ErrClosed
	}

	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}

	return blob.Reader()
}

func (m *localFilesystemBlobAccessMethod) getBlob() (blobaccess.BlobAccess, error) {
	if m.blobAccess == nil {
		data, err := m.base.GetBlob(m.spec.LocalReference)
		if err != nil {
			return nil, err
		}
		m.blobAccess = blobaccess.ForDataAccess(blobaccess.BLOB_UNKNOWN_DIGEST, blobaccess.BLOB_UNKNOWN_SIZE, m.MimeType(), data)
	}
	return m.blobAccess, m.err
}

func (m *localFilesystemBlobAccessMethod) Get() ([]byte, error) {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return nil, accessio.ErrClosed
	}

	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return blob.Get()
}

func (m *localFilesystemBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}

func (m *localFilesystemBlobAccessMethod) Close() error {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return accessio.ErrClosed
	}

	m.closed = true
	if m.blobAccess != nil {
		err := m.blobAccess.Close()
		return err
	}
	return nil
}
