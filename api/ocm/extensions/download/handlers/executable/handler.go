package executable

import (
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/compression"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
)

type Handler struct{}

func init() {
	h := &Handler{}
	download.Register(h, download.ForCombi(resourcetypes.OCM_PLUGIN, mime.MIME_OCTET))
	download.Register(h, download.ForCombi(resourcetypes.OCM_PLUGIN, mime.MIME_GZIP))
	download.Register(h, download.ForCombi(resourcetypes.EXECUTABLE, mime.MIME_OCTET))
	download.Register(h, download.ForCombi(resourcetypes.EXECUTABLE, mime.MIME_GZIP))
}

func wrapErr(err error, racc cpi.ResourceAccess) error {
	if err == nil {
		return nil
	}
	m := racc.Meta()
	return errors.Wrapf(err, "resource %s/%s%s", m.GetName(), m.GetVersion(), m.ExtraIdentity.String())
}

func (_ Handler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	rd, err := cpi.GetResourceReader(racc)
	if err != nil {
		return true, "", wrapErr(err, racc)
	}
	defer rd.Close()

	r, _, err := compression.AutoDecompress(rd)
	if err != nil {
		return true, "", err
	}
	if path == "" {
		path = racc.Meta().GetName()
	}

	t := ""
	if ok, err := vfs.Exists(fs, path); err == nil && ok {
		t = path
		path += ".new"
	}
	file, err := fs.OpenFile(path, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0o660)
	if err != nil {
		return true, "", wrapErr(errors.Wrapf(err, "creating target file %q", path), racc)
	}
	n, err := io.Copy(file, r)
	file.Close()
	if err == nil {
		if t != "" {
			err = fs.Remove(t)
			if err == nil {
				err = vfs.CopyFile(fs, path, fs, t)
			}
			if err == nil {
				err = fs.Remove(path)
			}
			if err == nil {
				path = t
			} else {
				p.Printf("cannot replace existing target file %s -> downloaded to %s\n", t, path)
			}
		}
		p.Printf("%s: %d byte(s) written\n", path, n)
		fs.Chmod(path, 0o755)
	} else {
		fs.Remove(path)
	}
	return true, path, wrapErr(err, racc)
}
