package plugin

import (
	"encoding/json"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/iotools"
)

type AccessDataWriter struct {
	plugin  Plugin
	creds   json.RawMessage
	accspec json.RawMessage
}

func NewAccessDataWriter(p Plugin, creds, accspec json.RawMessage) *AccessDataWriter {
	return &AccessDataWriter{p, creds, accspec}
}

func (d *AccessDataWriter) WriteTo(w accessio.Writer) (int64, digest.Digest, error) {
	dw := iotools.NewDefaultDigestWriter(accessio.NopWriteCloser(w))
	err := d.plugin.Get(dw, d.creds, d.accspec)
	if err != nil {
		return blobaccess.BLOB_UNKNOWN_SIZE, blobaccess.BLOB_UNKNOWN_DIGEST, err
	}
	return dw.Size(), dw.Digest(), nil
}
