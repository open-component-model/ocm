package transfer_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/oci/testhelper"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	ocictf "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	storagecontext "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const BASEURL = "baseurl.io"

func FakeOCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return BASEURL
}

var _ = Describe("disable upload", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()

		FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env.Builder)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})

		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers ctf with upload", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 artifact[ociImage](ocm/value:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())

		ctf := Must(ctfocm.Open(env, accessobj.ACC_READONLY, OUT, 0, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "version")

		res := Must(cv.GetResource(metav1.NewIdentity("artifact")))

		acc := Must(res.Access())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		Expect(acc.Describe(env.OCMContext())).To(Equal("OCI artifact " + BASEURL + "/" + OCINAMESPACE + ":" + OCIVERSION + "@sha256:" + D_OCIMANIFEST1))
	})

	It("transfers ctf without upload", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--disable-uploads", "--copy-resources", ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
standard blob upload handlers are disabled.
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 artifact[ociImage](ocm/value:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())

		ctf := Must(ctfocm.Open(env, accessobj.ACC_READONLY, OUT, 0, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "version")

		res := Must(cv.GetResource(metav1.NewIdentity("artifact")))

		acc := Must(res.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))
	})
})
