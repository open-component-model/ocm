package plugin

import (
	"encoding/json"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/iotools"
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
		return accessio.BLOB_UNKNOWN_SIZE, accessio.BLOB_UNKNOWN_DIGEST, err
	}
	return dw.Size(), dw.Digest(), nil
}
