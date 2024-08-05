package builder

const T_OCIPLATFORM = "platform consumer"

func (b *Builder) Platform(os string, arch string) {
	b.expect(b.oci_platform, T_OCIMANIFEST, func() bool { return b.oci_artacc.IsManifest() })

	b.oci_platform.OS = os
	b.oci_platform.Architecture = arch
}
