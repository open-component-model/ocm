// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package plugin_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const PLUGIN = "test"

const ARCH = "ctf"
const OUT = "/tmp/res"
const COMP = "github.com/mandelsoft/comp"
const VERS = "1.0.0"
const PROVIDER = "mandelsoft"
const RSCTYP = "TestArtefact"
const MEDIA = "text/plain"

const REPOTYPE = "test/v1"
const ACCTYPE = "test/v1"
const REPO = "plugin"
const CONTENT = "some test content\n"
const HINT = "given"

type RepoSpec struct {
	runtime.ObjectVersionedType
	Path string `json:"path"`
}

func NewRepoSpec(path string) *RepoSpec {
	return &RepoSpec{
		ObjectVersionedType: runtime.ObjectVersionedType{Type: REPOTYPE},
		Path:                path,
	}
}

type AccessSpec struct {
	runtime.ObjectVersionedType
	Path       string `json:"path"`
	MediaType  string `json:"mediaType"`
	Repository string `json:"repo"`
}

func NewAccessSpec(media, path, repo string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.ObjectVersionedType{Type: ACCTYPE},
		MediaType:           media,
		Path:                path,
		Repository:          repo,
	}
}

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var repodir string
	var env *Builder

	var accessSpec = NewAccessSpec(MEDIA, "given", REPO)
	var repoSpec = NewRepoSpec(REPO)

	BeforeEach(func() {
		repodir = Must(os.MkdirTemp(os.TempDir(), "uploadtest-*"))

		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())

		ctx.ConfigContext().ApplyConfig(config.New(PLUGIN, []byte(fmt.Sprintf(`{"root": "`+repodir+`"}`))), "plugin config")
		registry.RegisterExtensions()
	})

	AfterEach(func() {
		env.Cleanup()
		os.RemoveAll(repodir)
	})

	It("registers access methods", func() {

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERS, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", VERS, RSCTYP, metav1.LocalRelation, func() {
						env.Hint(HINT)
						env.BlobStringData(MEDIA, CONTENT)
						//env.Access(NewAccessSpec(MEDIA, "given", "dummy"))
					})
				})
			})
		})

		repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "source repo")

		cv := Must(repo.LookupComponentVersion(COMP, VERS))
		defer Close(cv, "source version")

		MustFailWithMessage(plugincacheattr.RegisterBlobHandler(env.OCMContext(), "test", "", RSCTYP, "", []byte("{}")),
			"plugin test: path missing in repository spec",
		)
		repospec := Must(json.Marshal(repoSpec))
		MustBeSuccessful(plugincacheattr.RegisterBlobHandler(env.OCMContext(), "test", "", RSCTYP, "", repospec))

		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer Close(tgt, "target repo")

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, Must(standard.New(standard.ResourcesByValue()))))
		Expect(env.DirExists(OUT)).To(BeTrue())

		Expect(vfs.FileExists(osfs.New(), filepath.Join(repodir, REPO, HINT))).To(BeTrue())

		tcv := Must(tgt.LookupComponentVersion(COMP, VERS))
		defer Close(tcv, "target version")

		r := Must(tcv.GetResourceByIndex(0))
		a := Must(r.Access())

		var spec AccessSpec
		MustBeSuccessful(json.Unmarshal(Must(json.Marshal(a)), &spec))
		Expect(spec).To(Equal(*accessSpec))

		m := Must(a.AccessMethod(tcv))
		defer Close(m, "method")

		Expect(string(Must(m.Get()))).To(Equal(CONTENT))
	})
})
