package builder

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

const T_OCICONFIG = "oci config"

type ociConfig struct {
	base
	blob blobaccess.BlobAccess
}

func (r *ociConfig) Type() string {
	return T_OCICONFIG
}

func (r *ociConfig) Set() {
	r.Builder.blob = &r.blob
}

func (r *ociConfig) Close() error {
	if r.blob == nil {
		return errors.Newf("config blob required")
	}
	m := r.Builder.oci_artacc.ManifestAccess()
	err := m.AddBlob(r.blob)
	if err != nil {
		return errors.Newf("cannot add config blob: %s", err)
	}
	d := artdesc.DefaultBlobDescriptor(r.blob)
	m.GetDescriptor().Config = *d
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Config(f ...func()) {
	b.expect(b.oci_artacc, T_OCIMANIFEST, func() bool { return b.oci_artacc.IsManifest() })
	b.configure(&ociConfig{}, f)
}
