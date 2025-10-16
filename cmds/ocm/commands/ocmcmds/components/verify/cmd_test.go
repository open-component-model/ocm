package verify_test

import (
	"bytes"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/resolvers"
	. "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH       = "/tmp/ctf"
	PROVIDER   = "mandelsoft"
	VERSION    = "v1"
	COMPONENTA = "github.com/mandelsoft/test"
	COMPONENTB = "github.com/mandelsoft/ref"
	OUT        = "/tmp/res"
	OCIPATH    = "/tmp/oci"
	OCIHOST    = "alias"
)

const (
	SIGNATURE = "test"
	SIGN_ALGO = rsa.Algorithm
)

const (
	PUBKEY  = "/tmp/pub"
	PRIVKEY = "/tmp/priv"
)

const (
	S_TESTDATA = "testdata"
	D_TESTDATA = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
)

const (
	S_OTHERDATA = "otherdata"
	D_OTHERDATA = "54b8007913ec5a907ca69001d59518acfd106f7b02f892eabf9cae3f8b2414b4"
)

const VERIFIED_FILE = "verified.yaml"

const (
	D_COMPONENTA = "01de99400030e8336020059a435cea4e7fe8f21aad4faf619da882134b85569d"
	D_COMPONENTB = "5f416ec59629d6af91287e2ba13c6360339b6a0acf624af2abd2a810ce4aefce"
)

var substitutions = Substitutions{
	"test": D_COMPONENTA,
	"r0":   D_TESTDATA,
	"r1":   DS_OCIMANIFEST1.Value,
	"r2":   DS_OCIMANIFEST2.Value,
	"ref":  D_COMPONENTB,
	"rb0":  D_OTHERDATA,
}

var _ = Describe("access method", func() {
	var (
		env *TestEnv
		log logging.Logger
	)

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	DefaultContext := ocm.DefaultContext()

	BeforeEach(func() {
		env = NewTestEnv()
		log = env.Logger()
		data, err := rsa.KeyData(pub)
		Expect(err).To(Succeed())
		Expect(vfs.WriteFile(env.FileSystem(), PUBKEY, data, os.ModePerm)).To(Succeed())
		data, err = rsa.KeyData(priv)
		Expect(err).To(Succeed())
		Expect(vfs.WriteFile(env.FileSystem(), PRIVKEY, data, os.ModePerm)).To(Succeed())

		FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env.Builder)
			OCIManifest2(env.Builder)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENTA, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, S_TESTDATA)
					})
					env.Resource("value", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
						)
						env.Label("transportByValue", true)
					})
					env.Resource("ref", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE2, OCIVERSION)),
						)
					})
				})
			})
			env.Component(COMPONENTB, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("otherdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, S_OTHERDATA)
					})
					env.Reference("ref", COMPONENTA, VERSION)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Prepare := func() {
		session := datacontext.NewSession()
		defer session.Close()

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		session.AddCloser(src)
		resolver := resolvers.NewCompoundResolver(src)

		cv := Must(resolver.LookupComponentVersion(COMPONENTB, VERSION))
		session.AddCloser(cv)

		opts := NewOptions(
			Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
			Resolver(resolver),
			PrivateKey(SIGNATURE, priv),
			Update(), VerifyDigests(),
		)
		Expect(opts.Complete(DefaultContext)).To(Succeed())

		dig := Must(Apply(nil, nil, cv, opts))
		log.Info("dig result", "dig", dig.String())
		Expect(dig.Value).To(Equal(D_COMPONENTB))
	}

	It("verifies transport archive", func() {
		buf := bytes.NewBuffer(nil)

		Prepare()

		session := datacontext.NewSession()
		defer session.Close()

		Expect(env.CatchOutput(buf).Execute("verify", "components", "-V", "-s", SIGNATURE, "-k", PUBKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(Succeed())

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${test}[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest SHA-256:${rb0}[genericBlobDigest/v1]
successfully verified github.com/mandelsoft/ref:v1 (digest SHA-256:${ref})
`, substitutions))
	})

	Context("verified store", func() {
		It("signs transport archive", func() {
			Prepare()

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("verify", "components", "--verified", VERIFIED_FILE, "-s", SIGNATURE, "-k", PUBKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${test}[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest SHA-256:${rb0}[genericBlobDigest/v1]
successfully verified github.com/mandelsoft/ref:v1 (digest SHA-256:${ref})
`, substitutions))

			Expect(Must(env.FileExists(VERIFIED_FILE))).To(BeTrue())

			store := Must(NewVerifiedStore(VERIFIED_FILE, env.FileSystem()))

			CheckStore(store, common.NewNameVersion(COMPONENTA, VERSION))
			CheckStore(store, common.NewNameVersion(COMPONENTB, VERSION))
		})
	})
})

func CheckStore(store VerifiedStore, ve common.VersionedElement) {
	e := store.Get(ve)
	ExpectWithOffset(1, e).NotTo(BeNil())
	ExpectWithOffset(1, common.VersionedElementKey(e)).To(Equal(common.VersionedElementKey(ve)))
}
