// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual_test

import (
	"path"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/virtual"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
)

var _ = Describe("virtual repo", func() {
	var env *TestEnv
	var repo ocm.Repository

	// ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, accessio.ALLOC_REALM))

	BeforeEach(func() {
		env = NewTestEnv(TestData())
		acc := Must(NewAccess(Must(projectionfs.New(env, "testdata"))))
		repo = virtual.NewRepository(env.OCMContext(), acc)
	})

	AfterEach(func() {
		MustBeSuccessful(repo.Close())
		env.Cleanup()
	})

	It("handles list", func() {
		lister := repo.ComponentLister()
		Expect(lister).NotTo(BeNil())
		names := Must(lister.GetComponents("", true))
		Expect(names).To(ConsistOf([]string{"acme.org/component", "acme.org/component/ref"}))
	})

	It("handles get", func() {
		comp := Must(repo.LookupComponent("acme.org/component"))
		defer Close(comp, "component")
		Expect(comp.ListVersions()).To(ConsistOf([]string{"v1.0.0"}))
		Expect(comp.HasVersion("v1.0.0")).To(BeTrue())
		Expect(comp.HasVersion("v1.0.1")).To(BeFalse())
		vers := Must(comp.LookupVersion("v1.0.0"))
		defer Close(vers, "version")
		r := Must(vers.GetResourceByIndex(0))
		data := Must(ocmutils.GetResourceData(r))
		Expect(string(data)).To(Equal("my test data\n"))

		a := Must(r.Access())
		Expect(a.GetInexpensiveContentVersionIdentity(vers)).To(Equal("sha256:2fdeb101f225dad71efd2dadb92b5aa422169f1884eecb81abdd988d77b68466"))
	})
})

////////////////////////////////////////////////////////////////////////////////

type Index = virtual.Index[interface{}]

type Access struct {
	fs    vfs.FileSystem
	index *Index
}

func NewAccess(fs vfs.FileSystem) (*Access, error) {
	a := &Access{
		fs:    fs,
		index: virtual.NewIndex[interface{}](),
	}

	list, err := vfs.ReadDir(fs, "descriptors")
	if err != nil {
		return nil, err
	}
	for _, e := range list {
		data, err := vfs.ReadFile(fs, path.Join("descriptors", e.Name()))
		if err != nil {
			return nil, err
		}
		cd, err := compdesc.Decode(data)
		if err != nil {
			return nil, err
		}
		err = a.index.Add(cd, nil)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *Access) ComponentLister() cpi.ComponentLister {
	return a.index
}

func (a *Access) ExistsComponentVersion(name string, version string) (bool, error) {
	e := a.index.Get(name, version)
	return e != nil, nil
}

func (a *Access) ListVersions(comp string) ([]string, error) {
	return a.index.GetVersions(comp), nil
}

func (a *Access) GetDescriptor(comp, version string) *compdesc.ComponentDescriptor {
	return a.index.Get(comp, version).CD()

}
func (a *Access) GetComponentVersion(comp, version string) (virtual.VersionAccess, error) {
	cd := a.GetDescriptor(comp, version)
	if cd == nil {
		// test is readonly mode
		return nil, errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, common.NewNameVersion(comp, version).String())
	}
	return &VersionAccess{a, cd.GetName(), cd.GetVersion(), cd.Copy()}, nil
}

func (a *Access) Close() error {
	return nil
}

var _ virtual.Access = (*Access)(nil)

type VersionAccess struct {
	access *Access
	comp   string
	vers   string
	desc   *compdesc.ComponentDescriptor
}

func (v *VersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	return v.desc
}

func (v *VersionAccess) GetBlob(name string) (cpi.DataAccess, error) {
	p := path.Join("blobs", name)

	if ok, err := vfs.FileExists(v.access.fs, p); !ok || err != nil {
		return nil, vfs.ErrNotExist
	}
	return accessio.DataAccessForFile(v.access.fs, p), nil
}

func (v *VersionAccess) AddBlob(blob cpi.BlobAccess) (string, error) {
	return "", accessio.ErrReadOnly
}

func (v *VersionAccess) Update() error {
	if v.desc.GetName() != v.comp || v.desc.GetVersion() != v.vers {
		return errors.ErrInvalid(cpi.KIND_COMPONENTVERSION, common.VersionedElementKey(v.desc).String())
	}
	if !reflect.DeepEqual(v.desc, v.access.GetDescriptor(v.comp, v.vers)) {
		return accessio.ErrReadOnly
	}
	return nil
}

func (v *VersionAccess) Close() error {
	return nil
}

func (v *VersionAccess) IsReadOnly() bool {
	return true
}

func (v *VersionAccess) GetInexpensiveContentVersionIdentity(a cpi.AccessSpec) string {
	switch a.GetKind() { //nolint:gocritic // to be extended
	case localblob.Type:
		blob, err := v.GetBlob(a.(*localblob.AccessSpec).LocalReference)
		if err != nil {
			return ""
		}
		defer blob.Close()
		dig, err := accessio.Digest(blob)
		if err != nil {
			return ""
		}
		return dig.String()
	}
	return ""
}

var _ virtual.VersionAccess = (*VersionAccess)(nil)
