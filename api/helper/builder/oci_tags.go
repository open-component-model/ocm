package builder

func (b *Builder) Tags(tags ...string) {
	b.expect(b.oci_tags, T_OCIARTIFACT)
	*b.oci_tags = append(*b.oci_tags, tags...)
}
