package add_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/oci/testhelper"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/defaultmerge"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

const (
	OCIPATH    = "/tmp/oci"
	OCIHOST    = "alias"
	ARCH       = "/tmp/ctf"
	LOOKUP     = "/tmp/lookup"
	PROVIDER   = "mandelsoft"
	VERSION    = "v1"
	COMPONENT  = "github.com/mandelsoft/test"
	COMPONENT2 = "github.com/mandelsoft/test2"
	OUT        = "/tmp/res"
)

func CheckComponent(env *TestEnv, handler func(ocm.Repository), tests ...func(cv ocm.ComponentVersionAccess)) {
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
	data := Must(ocmutils.GetResourceData(r))
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

	for _, t := range tests {
		t(cv)
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
		Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "testdata/component-constructor.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil)
	})

	It("creates ctf and adds component (deprecated)", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "testdata/component-constructor-old.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil)
	})

	It("creates ctf and adds components", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/component-constructor.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil)
	})

	It("creates ctf and adds components without digests", func() {
		Expect(env.Execute("add", "c", "--skip-digest-generation", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/component-constructor.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil, noDigest("data"), noDigest("text"))
	})
	It("creates ctf and adds components without digest for one resource", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "--version", "1.0.0", "testdata/component-constructor-skip.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		CheckComponent(env, nil, noDigest("data", false), noDigest("text"))
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
		var rname string // repo name
		_ = ldesc

		BeforeEach(func() {
			rname = FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

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
								ociartifact.New(oci.StandardOCIRef(rname, OCINAMESPACE, OCIVERSION)),
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
			Expect(env.Execute("add", "c", "-fcCV", "--lookup", LOOKUP, "--file", ARCH, "testdata/component-constructor.yaml")).To(Succeed())
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

func noDigest(name string, skips ...bool) func(cv ocm.ComponentVersionAccess) {
	skip := general.OptionalDefaultedBool(true, skips...)
	return func(cv ocm.ComponentVersionAccess) {
		r := MustWithOffset(1, Calling(cv.GetResource(metav1.Identity{"name": name})))
		if skip {
			ExpectWithOffset(1, r.Meta().Digest).To(BeNil())
		} else {
			ExpectWithOffset(1, r.Meta().Digest).NotTo(BeNil())
		}
	}
}
