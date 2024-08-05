package builder

func (b *Builder) Annotation(name, value string) {
	b.expect(b.oci_annofunc, T_OCIARTIFACT+" or "+T_OCIARTIFACTSET)
	b.oci_annofunc(name, value)
}
