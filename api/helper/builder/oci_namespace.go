package builder

import (
	"ocm.software/ocm/api/oci"
)

const T_OCINAMESPACE = "oci namespace"

type ociNamespace struct {
	base
	kind string
	oci.NamespaceAccess
	annofunc func(name, value string)
}

func (r *ociNamespace) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_OCINAMESPACE
}

func (r *ociNamespace) Set() {
	r.Builder.oci_nsacc = r.NamespaceAccess
	r.Builder.oci_annofunc = r.annofunc
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Namespace(name string, f ...func()) {
	b.expect(b.oci_repo, T_OCIREPOSITORY)
	r, err := b.oci_repo.LookupNamespace(name)
	b.failOn(err)
	b.configure(&ociNamespace{NamespaceAccess: r, kind: T_OCIARTIFACTSET}, f)
}
