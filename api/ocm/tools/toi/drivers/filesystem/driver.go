package filesystem

import (
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	"sigs.k8s.io/yaml"
)

const OptionTargetPath = "TARGET_PATH"

// Driver is capable of running Docker invocation images using Docker itself.
type Driver struct {
	config     map[string]string
	Simulate   bool
	TargetPath string
	Filesystem vfs.FileSystem
}

var _ install.Driver = (*Driver)(nil)

func New(fs vfs.FileSystem) install.Driver {
	if fs == nil {
		fs = osfs.New()
	}
	return &Driver{Filesystem: fs}
}

// SetConfig sets Docker driver configuration.
func (d *Driver) SetConfig(settings map[string]string) error {
	if settings != nil {
		d.TargetPath = settings[OptionTargetPath]
	}

	if d.TargetPath == "" {
		d.TargetPath = "toi"
	}
	d.config = settings
	return nil
}

func (d *Driver) Exec(op *install.Operation) (*install.OperationResult, error) {
	if d.Simulate {
		return nil, nil
	}

	err := d.Filesystem.MkdirAll(d.TargetPath, 0o700)
	if err != nil {
		return nil, errors.Wrapf(err, "creating target path")
	}

	var finalize finalizer.Finalizer
	defer finalize.Finalize()

	for k, v := range op.Files {
		n := vfs.Join(d.Filesystem, d.TargetPath, install.Inputs, k)
		err := d.Filesystem.MkdirAll(vfs.Dir(d.Filesystem, n), 0o700)
		if err != nil {
			return nil, errors.Wrapf(err, "creating directory for file %q", k)
		}
		r, err := v.Reader()
		if err != nil {
			return nil, errors.Wrapf(err, "reading data for %q", k)
		}
		finalize.Close(r)
		file, err := d.Filesystem.OpenFile(n, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0o600)
		if err != nil {
			return nil, errors.Wrapf(err, "writing file %q", n)
		}
		finalize.Close(file)
		_, err = io.Copy(file, r)
		if err != nil {
			return nil, errors.Wrapf(err, "writing %q", n)
		}
		finalize.Finalize()
	}
	props := map[string]string{}
	props["image"] = op.Image.String()
	props["componentVersion"] = op.ComponentVersion
	props["action"] = op.Action
	data, err := yaml.Marshal(props)
	if err != nil {
		return nil, errors.Wrapf(err, "writing operation properties")
	}
	vfs.WriteFile(d.Filesystem, vfs.Join(d.Filesystem, d.TargetPath, "properties"), data, 0o600)
	return &install.OperationResult{}, nil
}
