package artifactset

import (
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

func init() {
	h := &repospechandler{}
	cpi.RegisterRepositorySpecHandler(h, "")
	cpi.RegisterRepositorySpecHandler(h, Type)
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	path := u.Info
	if u.Info == "" {
		if u.Host == "" || u.Type == "" {
			return nil, nil
		}
		path = u.Host
	}
	fs := vfsattr.Get(ctx)

	hint, f := accessobj.MapType(u.TypeHint, Type, accessio.FormatDirectory, false)
	if !u.CreateIfMissing {
		hint = ""
	}

	create, ok, err := accessobj.CheckFile(Type, hint, accessio.TypeForTypeSpec(u.Type) == Type, path, fs, ArtifactSetDescriptorFileName)
	if err == nil && !ok {
		create, ok, err = accessobj.CheckFile(Type, hint, accessio.TypeForTypeSpec(u.Type) == Type, path, fs, OCIArtifactSetDescriptorFileName)
	}

	if !ok || err != nil {
		return nil, err
	}

	mode := accessobj.ACC_WRITABLE
	createHint := accessio.FormatNone
	if create {
		mode |= accessobj.ACC_CREATE
		createHint = f
	}
	return NewRepositorySpec(mode, path, createHint, accessio.PathFileSystem(fs))
}
