package loader

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	"ocm.software/ocm/api/tech/helm"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/iotools"
)

type Loader interface {
	ChartArchive() (blobaccess.BlobAccess, error)
	ChartArtefactSet() (blobaccess.BlobAccess, error)
	Chart() (*chart.Chart, error)
	Provenance() ([]byte, error)

	Close() error
}

type nopCloser = iotools.NopCloser

type vfsLoader struct {
	nopCloser
	path string
	fs   vfs.FileSystem
}

func VFSLoader(path string, fss ...vfs.FileSystem) Loader {
	return &vfsLoader{
		path: path,
		fs:   utils.FileSystem(fss...),
	}
}

func (l *vfsLoader) ChartArchive() (blobaccess.BlobAccess, error) {
	if ok, err := vfs.IsFile(l.fs, l.path); !ok || err != nil {
		return nil, err
	}
	return file.BlobAccess(helm.ChartMediaType, l.path, l.fs), nil
}

func (l *vfsLoader) ChartArtefactSet() (blobaccess.BlobAccess, error) {
	return nil, nil
}

func (l *vfsLoader) Chart() (*chart.Chart, error) {
	return Load(l.path, l.fs)
}

func (l *vfsLoader) Provenance() ([]byte, error) {
	prov := l.path + ".prov"
	if ok, err := vfs.FileExists(l.fs, prov); !ok || err != nil {
		return nil, err
	}
	return vfs.ReadFile(l.fs, prov)
}

////////////////////////////////////////////////////////////////////////////////

func Load(name string, fs vfs.FileSystem) (*chart.Chart, error) {
	fi, err := fs.Stat(name)
	if err != nil {
		return nil, errors.Wrapf(err, "%q not found", name)
	}
	if fi.IsDir() {
		c, err := LoadDir(fs, name)
		return c, errors.Wrapf(err, "cannot load chart %q", name)
	}
	file, err := fs.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open chart archive%q", name)
	}
	defer file.Close()
	c, err := loader.LoadArchive(file)
	return c, errors.Wrapf(err, "cannot load chart from %q", name)
}
