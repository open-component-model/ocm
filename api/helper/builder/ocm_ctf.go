package builder

import (
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const T_OCM_CTF = "ocm common transport format"

func (b *Builder) OCMCommonTransport(path string, fmt accessio.FileFormat, f ...func()) {
	r, err := ctf.Open(b.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0o777, fmt, accessio.PathFileSystem(b.FileSystem()))
	b.failOn(err)
	b.configure(&ocmRepository{Repository: r, kind: T_OCM_CTF}, f)
}
