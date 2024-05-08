package example

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path"
	"reflect"
	"sync"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/virtual"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////

func NewRepository(ctx cpi.ContextProvider, fs vfs.FileSystem, readonly bool, path ...string) (cpi.Repository, error) {
	var err error

	p := utils.Optional(path...)
	if p != "" && p != "/" {
		fs, err = projectionfs.New(fs, p)
		if err != nil {
			return nil, err
		}
	}
	acc, err := NewAccess(fs, readonly)
	if err != nil {
		return nil, err
	}
	return virtual.NewRepository(ctx.OCMContext(), acc), nil
}

type Index = virtual.Index[string]

type Access struct {
	lock     sync.Mutex
	readonly bool
	fs       vfs.FileSystem
	index    *Index
}

func NewAccess(fs vfs.FileSystem, readonly bool) (*Access, error) {
	a := &Access{
		readonly: readonly,
		fs:       fs,
	}
	err := a.Reset()
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Access) IsReadOnly() bool {
	return a.readonly
}

func (a *Access) SetReadOnly() {
	a.readonly = true
}

func (a *Access) Reset() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.index = virtual.NewIndex[string]()

	list, err := vfs.ReadDir(a.fs, "descriptors")
	if err != nil {
		return err
	}
	for _, e := range list {
		p := path.Join("descriptors", e.Name())
		data, err := vfs.ReadFile(a.fs, p)
		if err != nil {
			return err
		}
		cd, err := compdesc.Decode(data)
		if err != nil {
			return err
		}
		err = a.index.Add(cd, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Access) ComponentLister() cpi.ComponentLister {
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.index
}

func (a *Access) ExistsComponentVersion(name string, version string) (bool, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	e := a.index.Get(name, version)
	return e != nil, nil
}

func (a *Access) ListVersions(comp string) ([]string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.index.GetVersions(comp), nil
}

func (a *Access) GetComponentVersion(comp, version string) (virtual.VersionAccess, error) {
	var cd *compdesc.ComponentDescriptor

	a.lock.Lock()
	defer a.lock.Unlock()

	i := a.index.Get(comp, version)
	if i == nil {
		if a.readonly {
			return nil, errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, common.NewNameVersion(comp, version).String())
		}
		cd = compdesc.New(comp, version)
		hash := sha256.Sum256([]byte(comp + ":" + version))
		err := a.index.Add(cd, path.Join("descriptors", hex.EncodeToString(hash[:])))
		if err != nil {
			return nil, err
		}
	} else {
		cd = i.CD()
	}
	return &VersionAccess{a, cd.GetName(), cd.GetVersion(), a.readonly, cd.Copy()}, nil
}

func (a *Access) Close() error {
	return nil
}

var _ virtual.Access = (*Access)(nil)

type VersionAccess struct {
	access   *Access
	comp     string
	vers     string
	readonly bool
	desc     *compdesc.ComponentDescriptor
}

func (v *VersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	return v.desc
}

func (v *VersionAccess) GetBlob(name string) (cpi.DataAccess, error) {
	p := path.Join("blobs", name)

	if ok, err := vfs.FileExists(v.access.fs, p); !ok || err != nil {
		return nil, vfs.ErrNotExist
	}
	return blobaccess.DataAccessForFile(v.access.fs, p), nil
}

func (v *VersionAccess) AddBlob(blob cpi.BlobAccess) (string, error) {
	if v.IsReadOnly() {
		return "", accessio.ErrReadOnly
	}
	d := blob.Digest()
	p := path.Join("blobs", d.Encoded())
	r, err := blob.Reader()
	if err != nil {
		return "", err
	}
	defer r.Close()
	w, err := v.access.fs.OpenFile(p, vfs.O_CREATE|vfs.O_RDWR, 0o600)
	if err != nil {
		return "", err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	if err != nil {
		return "", err
	}
	return d.Encoded(), nil
}

func (v *VersionAccess) Update() error {
	v.access.lock.Lock()
	defer v.access.lock.Unlock()

	if v.desc.GetName() != v.comp || v.desc.GetVersion() != v.vers {
		return errors.ErrInvalid(cpi.KIND_COMPONENTVERSION, common.VersionedElementKey(v.desc).String())
	}
	i := v.access.index.Get(v.comp, v.vers)
	if !reflect.DeepEqual(v.desc, i.CD()) {
		if v.IsReadOnly() {
			return accessio.ErrReadOnly
		}
		data, err := compdesc.Encode(v.desc)
		if err != nil {
			return err
		}
		v.access.index.Set(v.desc, i.Info())
		return vfs.WriteFile(v.access.fs, i.Info(), data, 0o600)
	}
	return nil
}

func (v *VersionAccess) Close() error {
	return v.Update()
}

func (v *VersionAccess) IsReadOnly() bool {
	return v.readonly || v.access.readonly
}

func (v *VersionAccess) SetReadOnly() {
	v.readonly = true
}

func (v *VersionAccess) GetInexpensiveContentVersionIdentity(a cpi.AccessSpec) string {
	switch a.GetKind() { //nolint:gocritic // to be extended
	case localblob.Type:
		blob, err := v.GetBlob(a.(*localblob.AccessSpec).LocalReference)
		if err != nil {
			return ""
		}
		defer blob.Close()
		dig, err := blobaccess.Digest(blob)
		if err != nil {
			return ""
		}
		return dig.String()
	}
	return ""
}

var _ virtual.VersionAccess = (*VersionAccess)(nil)
