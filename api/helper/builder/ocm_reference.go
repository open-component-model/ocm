package builder

import (
	"ocm.software/ocm/api/ocm/compdesc"
)

type ocmReference struct {
	base

	meta compdesc.Reference
}

const T_OCMREF = "reference"

func (r *ocmReference) Type() string {
	return T_OCMREF
}

func (r *ocmReference) Set() {
	r.Builder.ocm_meta = &r.meta.ElementMeta
	r.Builder.ocm_labels = &r.meta.ElementMeta.Labels
}

func (r *ocmReference) Close() error {
	return r.ocm_vers.SetReference(&r.meta)
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Reference(name, comp, vers string, f ...func()) {
	b.expect(b.ocm_vers, T_OCMVERSION)
	r := &ocmReference{}
	r.meta.Name = name
	r.meta.Version = vers
	r.meta.ComponentName = comp
	b.configure(r, f)
}
