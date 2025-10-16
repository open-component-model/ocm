package virtual

import (
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type localBlobAccessMethod struct {
	lock sync.Mutex
	data blobaccess.DataAccess
	spec *localblob.AccessSpec
}

var _ accspeccpi.AccessMethodImpl = (*localBlobAccessMethod)(nil)

func newLocalBlobAccessMethod(a *localblob.AccessSpec, data blobaccess.DataAccess) (*localBlobAccessMethod, error) {
	return &localBlobAccessMethod{
		spec: a,
		data: data,
	}, nil
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

	if m.data == nil {
		return blobaccess.ErrClosed
	}
	list := errors.ErrorList{}
	list.Add(m.data.Close())
	m.data = nil
	return list.Result()
}

func (m *localBlobAccessMethod) Reader() (io.ReadCloser, error) {
	return m.data.Reader()
}

func (m *localBlobAccessMethod) Get() (data []byte, ferr error) {
	return blobaccess.BlobData(m.data)
}

func (m *localBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
