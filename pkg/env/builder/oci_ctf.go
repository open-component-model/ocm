package builder

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
)

const T_OCI_CTF = "oci common transport format"

func (b *Builder) OCICommonTransport(path string, fmt accessio.FileFormat, f ...func()) {
	r, err := ctf.Open(b.OCMContext().OCIContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0o777, accessio.PathFileSystem(b.FileSystem()))
	b.failOn(err)
	b.configure(&ociRepository{Repository: r, kind: T_OCI_CTF}, f)
}
