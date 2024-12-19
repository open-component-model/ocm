package genericocireg

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
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
	refs := strings.Split(m.spec.LocalReference, ",")

	var (
		data blobaccess.DataAccess
		err  error
	)
	if len(refs) < 2 {
		_, data, err = m.namespace.GetBlobData(digest.Digest(m.spec.LocalReference))
		if err != nil {
			return nil, err
		}
	} else {
		data = &composedBlock{m, refs}
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

////////////////////////////////////////////////////////////////////////////////

type composedBlock struct {
	m    *localBlobAccessMethod
	refs []string
}

var _ blobaccess.DataAccess = (*composedBlock)(nil)

func (c *composedBlock) Get() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for _, ref := range c.refs {
		var finalize finalizer.Finalizer

		_, data, err := c.m.namespace.GetBlobData(digest.Digest(ref))
		if err != nil {
			return nil, err
		}
		finalize.Close(data)
		r, err := data.Reader()
		if err != nil {
			return nil, err
		}
		finalize.Close(r)
		_, err = io.Copy(buf, r)
		if err != nil {
			return nil, err
		}
		err = finalize.Finalize()
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (c *composedBlock) Reader() (io.ReadCloser, error) {
	return &composedReader{
		m:    c.m,
		refs: c.refs,
	}, nil
}

func (c *composedBlock) Close() error {
	return nil
}

type composedReader struct {
	lock   sync.Mutex
	m      *localBlobAccessMethod
	refs   []string
	reader io.ReadCloser
	data   blobaccess.DataAccess
}

func (c *composedReader) Read(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for {
		if c.reader != nil {
			n, err := c.reader.Read(p)

			if err == io.EOF {
				c.reader.Close()
				c.data.Close()
				c.refs = c.refs[1:]
				c.reader = nil
				c.data = nil
				// start new layer and return partial (>0) read before next layer is started
				err = nil
			}
			// return partial read (even a zero read if layer is not yet finished) or error
			if c.reader != nil || err != nil || n > 0 {
				return n, err
			}
			// otherwise, we can use the given buffer for the next layer

			// now, we have to check for a next succeeding layer.
			// This means to finish with the actual reader and continue
			// with the next one.
		}

		// If no more layers are available, report EOF.
		if len(c.refs) == 0 {
			return 0, io.EOF
		}

		ref := strings.TrimSpace(c.refs[0])
		_, c.data, err = c.m.namespace.GetBlobData(digest.Digest(ref))
		if err != nil {
			return 0, err
		}
		c.reader, err = c.data.Reader()
		if err != nil {
			return 0, err
		}
	}
}

func (c *composedReader) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.reader == nil && c.refs == nil {
		return os.ErrClosed
	}
	if c.reader != nil {
		c.reader.Close()
		c.data.Close()
		c.reader = nil
		c.refs = nil
	}
	return nil
}
