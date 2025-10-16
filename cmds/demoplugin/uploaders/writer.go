package uploaders

import (
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/iotools"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/demoplugin/accessmethods"
)

type writer = iotools.DigestWriter

type Writer struct {
	*writer
	file    *os.File
	path    string
	rename  bool
	name    string
	version string
	media   string
	spec    *accessmethods.AccessSpec
}

func NewWriter(file *os.File, path string, media string, rename bool, name, version string) *Writer {
	return &Writer{
		writer:  iotools.NewDefaultDigestWriter(file),
		file:    file,
		path:    path,
		rename:  rename,
		name:    name,
		version: version,
		media:   media,
	}
}

func (w *Writer) Close() error {
	err := w.writer.Close()
	if err == nil {
		n := w.path
		if w.rename {
			n = filepath.Join(os.TempDir(), n, common.DigestToFileName(w.writer.Digest()))
			err := os.Rename(w.file.Name(), n)
			if err != nil {
				return errors.Wrapf(err, "cannot rename %q to %q", w.file.Name(), n)
			}
		}
		w.spec = &accessmethods.AccessSpec{
			ObjectVersionedType: runtime.NewVersionedTypedObject(w.name, w.version),
			Path:                n,
			MediaType:           w.media,
		}
	}
	return err
}

func (w *Writer) Specification() ppi.AccessSpec {
	return w.spec
}
