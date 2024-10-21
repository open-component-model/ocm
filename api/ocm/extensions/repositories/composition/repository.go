package composition

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/virtual"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	common "ocm.software/ocm/api/utils/misc"
)

////////////////////////////////////////////////////////////////////////////////

func NewRepository(ctxp cpi.ContextProvider, names ...string) cpi.Repository {
	var repositories *Repositories

	ctx := datacontext.InternalContextRef(ctxp.OCMContext())
	name := general.Optional(names...)
	if name != "" {
		repositories = ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
		if repo := repositories.GetRepository(name); repo != nil {
			repo, _ = repo.Dup()
			return repo
		}
	}
	repo := virtual.NewRepository(ctx, NewAccess(name))
	if repositories != nil {
		repositories.SetRepository(name, repo)
		repo, _ = repo.Dup()
	}
	return repo
}

type Index = virtual.Index[common.NameVersion]

type Access struct {
	lock     sync.Mutex
	name     string
	index    *Index
	blobs    map[string]blobaccess.BlobAccess
	readonly bool
}

var _ virtual.Access = (*Access)(nil)

func NewAccess(name string) *Access {
	return &Access{
		name:  name,
		index: virtual.NewIndex[common.NameVersion](),
		blobs: map[string]blobaccess.BlobAccess{},
	}
}

func (a *Access) GetSpecification() cpi.RepositorySpec {
	return NewRepositorySpec(a.name)
}

func (a *Access) IsReadOnly() bool {
	return a.readonly
}

func (a *Access) SetReadOnly() {
	a.readonly = true
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
		cd = compdesc.New(comp, version)
	} else {
		cd = i.CD()
	}
	return &VersionAccess{a, cd.GetName(), cd.GetVersion(), a.IsReadOnly(), cd.Copy(), false}, nil
}

func (a *Access) GetBlob(name string) (blobaccess.BlobAccess, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	b := a.blobs[name]
	if b == nil {
		return nil, errors.ErrNotFound(blobaccess.KIND_BLOB, name)
	}
	return b.Dup()
}

func (a *Access) AddBlob(blob blobaccess.BlobAccess) (string, error) {
	digest := blob.Digest()
	if digest == blobaccess.BLOB_UNKNOWN_DIGEST {
		return "", fmt.Errorf("unknown digest")
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	b := a.blobs[digest.Encoded()]
	if b == nil {
		b, err := blob.Dup()
		if err != nil {
			return "", err
		}
		a.blobs[digest.Encoded()] = b
	}
	return digest.Encoded(), nil
}

func (a *Access) Close() error {
	list := errors.ErrorList{}
	for _, b := range a.blobs {
		list.Add(b.Close())
	}
	return list.Result()
}

var _ virtual.Access = (*Access)(nil)

type VersionAccess struct {
	access   *Access
	comp     string
	vers     string
	readonly bool
	desc     *compdesc.ComponentDescriptor
	new      bool
}

func (v *VersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	return v.desc
}

func (v *VersionAccess) GetBlob(name string) (cpi.DataAccess, error) {
	return v.access.GetBlob(name)
}

func (v *VersionAccess) AddBlob(blob cpi.BlobAccess) (string, error) {
	if v.readonly {
		return "", accessio.ErrReadOnly
	}
	return v.access.AddBlob(blob)
}

func (v *VersionAccess) Update() (bool, error) {
	v.access.lock.Lock()
	defer v.access.lock.Unlock()

	if v.readonly {
		return false, accessio.ErrReadOnly
	}
	if v.desc.GetName() != v.comp || v.desc.GetVersion() != v.vers {
		return false, errors.ErrInvalid(cpi.KIND_COMPONENTVERSION, common.VersionedElementKey(v.desc).String())
	}
	i := v.access.index.Get(v.comp, v.vers)
	if !reflect.DeepEqual(v.desc, i.CD()) {
		if v.new {
			err := v.access.index.Add(v.desc, i.Info())
			if err != nil {
				return false, err
			}
			v.new = false
		} else {
			v.access.index.Set(v.desc, i.Info())
		}
		return true, nil
	}
	return false, nil
}

func (v *VersionAccess) Close() error {
	_, err := v.Update()
	return err
}

func (v *VersionAccess) IsReadOnly() bool {
	return v.readonly
}

func (v *VersionAccess) SetReadOnly() {
	v.readonly = true
}

var _ virtual.VersionAccess = (*VersionAccess)(nil)
