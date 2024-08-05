package ctf

import (
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/utils/accessio"
)

func SupportedFormats() []accessio.FileFormat {
	return ctf.SupportedFormats()
}

func init() {
	h := &repospechandler{}
	cpi.RegisterRepositorySpecHandler(h, "")
	cpi.RegisterRepositorySpecHandler(h, ctf.Type)
	cpi.RegisterRepositorySpecHandler(h, ctf.AltType)
	for _, f := range SupportedFormats() {
		cpi.RegisterRepositorySpecHandler(h, string(f))
	}
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	if u.Info == "" {
		if u.Host == "" || u.Type == "" {
			return nil, nil
		}
	}
	spec, err := ctf.MapReference(ctx.OCIContext(), &oci.UniformRepositorySpec{
		Type:            u.Type,
		Host:            u.Host,
		Info:            u.Info,
		CreateIfMissing: u.CreateIfMissing,
		TypeHint:        u.TypeHint,
	})
	if err != nil || spec == nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
}
