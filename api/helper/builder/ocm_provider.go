package builder

import (
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

type ocmProvider struct {
	base

	provider compdesc.Provider
}

const T_OCMPROVIDER = "provider"

func (r *ocmProvider) Type() string {
	return T_OCMPROVIDER
}

func (r *ocmProvider) Set() {
	r.Builder.ocm_labels = &r.provider.Labels
}

func (r *ocmProvider) Close() error {
	r.ocm_vers.GetDescriptor().Provider = r.provider
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Provider(name string, f ...func()) {
	b.expect(b.ocm_vers, T_OCMVERSION)
	r := &ocmProvider{}
	r.provider.Name = metav1.ProviderName(name)
	b.configure(r, f)
}
