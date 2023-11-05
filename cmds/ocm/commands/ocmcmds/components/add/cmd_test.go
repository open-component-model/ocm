// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/defaultmerge"
	"github.com/open-component-model/ocm/pkg/mime"
)

const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"
const ARCH = "/tmp/ctf"
const LOOKUP = "/tmp/lookup"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const COMPONENT2 = "github.com/mandelsoft/test2"
const OUT = "/tmp/res"

func CheckComponent(env *TestEnv, handler func(ocm.Repository)) {
	repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
	defer Close(repo)
	cv := Must(repo.LookupComponentVersion("ocm.software/demo/test", "1.0.0"))
	defer Close(cv)
	cd := cv.GetDescriptor()

	var plabels metav1.Labels
	MustBeSuccessful(plabels.Set("city", "Karlsruhe"))

	var clabels metav1.Labels
	MustBeSuccessful(clabels.Set("purpose", "test"))

	var rlabels metav1.Labels
	MustBeSuccessful(rlabels.Set("city", "Karlsruhe", metav1.WithMerging(defaultmerge.ALGORITHM, defaultmerge.NewConfig(defaultmerge.MODE_INBOUND))))

	Expect(string(cd.Provider.Name)).To(Equal("ocm.software"))
	Expect(cd.Provider.Labels).To(Equal(plabels))
	Expect(cd.Labels).To(Equal(clabels))

	r := Must(cv.GetResource(metav1.Identity{"name": "data"}))
	data := Must(ocm.ResourceData(r))
	Expect(string(data)).To(Equal("!stringdata"))

	r = Must(cv.GetResource(metav1.Identity{"name": "text"}))
	Expect(r.Meta().Labels).To(Equal(rlabels))

	Expect(cv.GetDescriptor().References).To(Equal(compdesc.References{{
		ElementMeta: compdesc.ElementMeta{
			Name:    "ref",
			Version: VERSION,
		},
		ComponentName: COMPONENT2,
	}}))

	if handler != nil {
		handler(repo)
	}
}

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates ctf and adds component", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "testdata/component.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil)
	})

	It("creates ctf and adds components", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/components.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil)
	})

	Context("failures", func() {
		It("rejects adding duplicate components", func() {
			ExpectError(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/components-dup.yaml")).To(
				MatchError(`duplicate component identity "name"="ocm.software/demo/test","version"="1.0.0" (testdata/components-dup.yaml[1][2] and testdata/components-dup.yaml[1][1])`),
			)
		})
		It("rejects adding duplicate resources", func() {
			ExpectError(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/component-dup-res.yaml")).To(
				MatchError(`duplicate resource identity "name"="text" (testdata/component-dup-res.yaml[1][1] index 3 and 1)`),
			)
		})
		It("rejects adding duplicate source", func() {
			ExpectError(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/component-dup-src.yaml")).To(
				MatchError(`duplicate source identity "name"="source" (testdata/component-dup-src.yaml[1][1] index 2 and 1)`),
			)
		})
		It("rejects adding duplicate reference", func() {
			ExpectError(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/component-dup-ref.yaml")).To(
				MatchError(`duplicate reference identity "name"="ref","version"="v1" (testdata/component-dup-ref.yaml[1][1] index 2 and 1)`),
			)
		})
	})

	Context("with completion", func() {
		var ldesc *artdesc.Descriptor

		_ = ldesc

		BeforeEach(func() {
			FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				ldesc = OCIManifest1(env.Builder)
				OCIManifest2(env.Builder)
			})

			env.OCMCommonTransport(LOOKUP, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
						env.Resource("image", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
							)
						})
					})
				})
				env.Component(COMPONENT2, func() {
					env.Version(VERSION, func() {
						env.Reference("ref", COMPONENT, VERSION)
						env.Provider(PROVIDER)
					})
				})
			})
		})

		It("creates ctf and adds components", func() {
			Expect(env.Execute("add", "c", "-fcCV", "--lookup", LOOKUP, "--file", ARCH, "testdata/component.yaml")).To(Succeed())
			Expect(env.DirExists(ARCH)).To(BeTrue())
			CheckComponent(env, func(repo ocm.Repository) {
				cv := MustWithOffset(2, R(repo.LookupComponentVersion(COMPONENT, VERSION)))
				defer Close(cv)
				res := MustWithOffset(2, R(cv.GetResource(metav1.Identity{"name": "image"})))
				Expect(MustWithOffset(2, R(res.Access())).GetKind()).To(Equal(localblob.Type))
				Expect(MustWithOffset(2, R(res.Access())).GlobalAccessSpec(env.OCMContext()).GetKind()).To(Equal(ociartifact.Type))
			})
		})
	})
})
