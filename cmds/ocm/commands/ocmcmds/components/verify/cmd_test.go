package verify_test

import (
	"bytes"
	"os"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENTA = "github.com/mandelsoft/test"
const COMPONENTB = "github.com/mandelsoft/ref"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

const SIGNATURE = "test"
const SIGN_ALGO = rsa.Algorithm

const PUBKEY = "/tmp/pub"
const PRIVKEY = "/tmp/priv"

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
						env.BlobStringData(mime.MIME_TEXT, "testdata")
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
						env.BlobStringData(mime.MIME_TEXT, "otherdata")
					})
					env.Reference("ref", COMPONENTA, VERSION)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("sign component archive", func() {
		buf := bytes.NewBuffer(nil)
		digest := "5f416ec59629d6af91287e2ba13c6360339b6a0acf624af2abd2a810ce4aefce"

		session := datacontext.NewSession()
		defer session.Close()

		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
		Expect(err).To(Succeed())
		archcloser := session.AddCloser(src)
		resolver := ocm.NewCompoundResolver(src)

		cv, err := resolver.LookupComponentVersion(COMPONENTB, VERSION)
		Expect(err).To(Succeed())
		closer := session.AddCloser(cv)

		opts := NewOptions(
			Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
			Resolver(resolver),
			PrivateKey(SIGNATURE, priv),
			Update(), VerifyDigests(),
		)
		Expect(opts.Complete(DefaultContext)).To(Succeed())
		dig, err := Apply(nil, nil, cv, opts)
		Expect(err).To(Succeed())
		closer.Close()
		archcloser.Close()
		log.Info("dig result", "dig", dig.String())
		Expect(dig.Value).To(Equal(digest))

		Expect(env.CatchOutput(buf).Execute("verify", "components", "-V", "-s", SIGNATURE, "-k", PUBKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(Succeed())

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:c2d2dca275c33c1270dea6168a002d67c0e98780d7a54960758139ae19984bd7[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:01de99400030e8336020059a435cea4e7fe8f21aad4faf619da882134b85569d[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest SHA-256:54b8007913ec5a907ca69001d59518acfd106f7b02f892eabf9cae3f8b2414b4[genericBlobDigest/v1]
successfully verified github.com/mandelsoft/ref:v1 (digest SHA-256:` + digest + `)
`))
	})
})
