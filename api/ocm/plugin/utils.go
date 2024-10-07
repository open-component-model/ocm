package plugin

import (
	"encoding/json"
	"io"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/iotools"
)

type BlobAccessWriter struct {
	creds  json.RawMessage
	spec   json.RawMessage
	getter func(writer io.Writer, creds json.RawMessage, spec json.RawMessage) error
}

func NewAccessDataWriter(p Plugin, creds, spec json.RawMessage) *BlobAccessWriter {
	return &BlobAccessWriter{creds, spec, p.Get}
}

func NewInputDataWriter(p Plugin, dir string, creds, spec json.RawMessage) *BlobAccessWriter {
	return &BlobAccessWriter{
		creds:  creds,
		spec:   spec,
		getter: func(w io.Writer, creds, spec json.RawMessage) error { return p.GetInputBlob(w, dir, creds, spec) },
	}
}

func (d *BlobAccessWriter) WriteTo(w accessio.Writer) (int64, digest.Digest, error) {
	dw := iotools.NewDefaultDigestWriter(accessio.NopWriteCloser(w))
	err := d.getter(dw, d.creds, d.spec)
	if err != nil {
		return blobaccess.BLOB_UNKNOWN_SIZE, blobaccess.BLOB_UNKNOWN_DIGEST, err
	}
	return dw.Size(), dw.Digest(), nil
}
