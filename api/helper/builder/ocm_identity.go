package builder

const T_OCMMETA = "element with metadata"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ExtraIdentity(name string, value string) {
	b.expect(b.ocm_meta, T_OCMMETA)

	b.ocm_meta.ExtraIdentity.Set(name, value)
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
