package uploaders

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/demoplugin/accessmethods"
	"ocm.software/ocm/cmds/demoplugin/common"
	"ocm.software/ocm/cmds/demoplugin/config"
)

const (
	NAME    = "demo"
	VERSION = "v1"
)

type TargetSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	Path string `json:"path"`
}

var types ppi.UploadFormats

func init() {
	decoder, err := runtime.NewDirectDecoder[runtime.TypedObject](&TargetSpec{})
	if err != nil {
		panic(err)
	}
	types = ppi.UploadFormats{NAME + runtime.VersionSeparator + VERSION: decoder}
}

type Uploader struct {
	ppi.UploaderBase
}

var _ ppi.Uploader = (*Uploader)(nil)

func New() ppi.Uploader {
	return &Uploader{
		UploaderBase: ppi.MustNewUploaderBase("demo", "upload temp files"),
	}
}

func (a *Uploader) Decoders() ppi.UploadFormats {
	return types
}

func (a *Uploader) ValidateSpecification(_ ppi.Plugin, spec ppi.UploadTargetSpec) (*ppi.UploadTargetSpecInfo, error) {
	var info ppi.UploadTargetSpecInfo
	my := spec.(*TargetSpec)

	if strings.HasPrefix(my.Path, "/") {
		return nil, fmt.Errorf("path must be relative (%s)", my.Path)
	}

	info.ConsumerId = credentials.ConsumerIdentity{
		cpi.ID_TYPE:            common.CONSUMER_TYPE,
		identity.ID_HOSTNAME:   "localhost",
		identity.ID_PATHPREFIX: my.Path,
	}
	return &info, nil
}

func (a *Uploader) Upload(_ context.Context, p ppi.Plugin, arttype, mediatype, hint, digest string, spec ppi.UploadTargetSpec, creds credentials.Credentials, reader io.Reader) (ppi.AccessSpecProvider, error) {
	var file *os.File
	var err error

	cfg, err := p.GetConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "can't get config for access method %s", mediatype)
	}

	root := os.TempDir()
	if cfg != nil && cfg.(*config.Config).Uploaders.Path != "" {
		root = cfg.(*config.Config).Uploaders.Path
		err := os.MkdirAll(root, 0o700)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create root dir")
		}
	}

	path := hint
	my := spec.(*TargetSpec)
	dir := root
	if my.Path != "" {
		root = filepath.Join(root, my.Path)
		if hint == "" {
			path = my.Path
			dir = filepath.Join(dir, path)
		} else {
			path = filepath.Join(my.Path, hint)
			dir = filepath.Join(dir, filepath.Dir(path))
		}
	}

	err = os.MkdirAll(dir, 0o700)
	if err != nil {
		return nil, err
	}

	if hint == "" {
		file, err = os.CreateTemp(root, "demo.*.blob")
	} else {
		file, err = os.OpenFile(filepath.Join(os.TempDir(), path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	}
	if err != nil {
		return nil, err
	}
	writer := NewWriter(file, path, mediatype, hint == "", accessmethods.NAME, accessmethods.VERSION)

	if _, err = io.Copy(writer, reader); err != nil {
		return nil, fmt.Errorf("cannot write to %q: %w", file.Name(), err)
	}

	return writer.Specification, nil
}
