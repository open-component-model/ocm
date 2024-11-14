package builder

import (
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

const T_OCMMETA = "element with metadata"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ExtraIdentity(name string, value string) {
	b.expect(b.ocm_meta, T_OCMMETA)

	b.ocm_meta.ExtraIdentity.Set(name, value)
}

func (b *Builder) ExtraIdentities(extras ...string) {
	b.expect(b.ocm_meta, T_OCMMETA)

	id := metav1.NewExtraIdentity(extras...)
	if b.ocm_meta.ExtraIdentity == nil {
		b.ocm_meta.ExtraIdentity = metav1.Identity{}
	}
	for k, v := range id {
		b.ocm_meta.ExtraIdentity.Set(k, v)
	}
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) RemoveExtraIdentity(name string) {
	b.expect(b.ocm_meta, T_OCMMETA)

	b.ocm_meta.ExtraIdentity.Remove(name)
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ClearExtraIdentities() {
	b.expect(b.ocm_meta, T_OCMMETA)

	b.ocm_meta.ExtraIdentity = nil
}
