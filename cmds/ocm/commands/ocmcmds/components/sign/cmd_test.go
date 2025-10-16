package sign_test

import (
	"bytes"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	. "ocm.software/ocm/api/ocm/testhelper"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	COMPARCH   = "/tmp/ca"
	ARCH       = "/tmp/ctf"
	ARCH2      = "/tmp/ctf2"
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
	D_COMPONENTA_V1 = "01de99400030e8336020059a435cea4e7fe8f21aad4faf619da882134b85569d"
	D_COMPONENTB_V1 = "5f416ec59629d6af91287e2ba13c6360339b6a0acf624af2abd2a810ce4aefce"
)

const VERIFIED_FILE = "verified.yaml"

var substitutionsV1 = Substitutions{
	"test":     D_COMPONENTA_V1,
	"r0":       D_TESTDATA,
	"r1":       DS_OCIMANIFEST1.Value,
	"r2":       DS_OCIMANIFEST2.Value,
	"ref":      D_COMPONENTB_V1,
	"rb0":      D_OTHERDATA,
	"normAlgo": compdesc.JsonNormalisationV1,
}

const (
	D_COMPONENTA_V2 = "10ac0b3a850e1f1becf56d5d45e9742fa0a91103d25ba93cc3a509f68797e90f"
	D_COMPONENTB_V2 = "1ae74420ef29436ad75133d81bceb59fa8ef1e2ce083a45b5f4baaec641a4266"
)

var substitutionsV2 = Substitutions{
	"test":     D_COMPONENTA_V2,
	"r0":       D_TESTDATA,
	"r1":       DS_OCIMANIFEST1.Value,
	"r2":       DS_OCIMANIFEST2.Value,
	"ref":      D_COMPONENTB_V2,
	"rb0":      D_OTHERDATA,
	"normAlgo": compdesc.JsonNormalisationV2,
}

const (
	D_COMPONENTA_V3 = D_COMPONENTA_V2
	D_COMPONENTB_V3 = "766f26b09237f9647714e85fac914f115d0b4c3277b01ec00cfeb3b50a68cde9"
)

var substitutionsV3 = Substitutions{
	"test":     D_COMPONENTA_V3,
	"r0":       D_TESTDATA,
	"r1":       DS_OCIMANIFEST1.Value,
	"r2":       DS_OCIMANIFEST2.Value,
	"ref":      D_COMPONENTB_V3,
	"rb0":      D_OTHERDATA,
	"normAlgo": compdesc.JsonNormalisationV3,
}

var _ = Describe("access method", func() {
	var env *TestEnv

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	BeforeEach(func() {
		env = NewTestEnv()
		data, err := rsa.KeyData(pub)
		Expect(err).To(Succeed())
		Expect(vfs.WriteFile(env.FileSystem(), PUBKEY, data, os.ModePerm)).To(Succeed())
		data, err = rsa.KeyData(priv)
		Expect(err).To(Succeed())
		Expect(vfs.WriteFile(env.FileSystem(), PRIVKEY, data, os.ModePerm)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("valid", func() {
		BeforeEach(func() {
			FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIManifest1(env.Builder)
				OCIManifest2(env.Builder)
			})
		})

		It("has digests", func() {
			prepareEnv(env, ARCH, ARCH)

			repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
			defer Close(repo, "repo")
			cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
			defer Close(cv, "cva")

			r := Must(cv.GetResource(metav1.NewIdentity("value")))
			Expect(r.Meta().Digest).To(Equal(DS_OCIMANIFEST1))

			r = Must(cv.GetResource(metav1.NewIdentity("ref")))
			Expect(r.Meta().Digest).To(Equal(DS_OCIMANIFEST2))
		})

		It("signs single component in component archive", func() {
			prepareEnv(env, ARCH, "")

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTA+":"+VERSION, "--normalization", compdesc.JsonNormalisationV1)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
  resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
  resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
  resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
successfully signed github.com/mandelsoft/test:v1 (digest SHA-256:${test})`,
				substitutionsV1),
			)

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPONENTA_V1))
		})

		It("signs transport archive", func() {
			prepareEnv(env, ARCH, ARCH)

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION, "--normalization", compdesc.JsonNormalisationV1)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${test}[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest SHA-256:${rb0}[genericBlobDigest/v1]
successfully signed github.com/mandelsoft/ref:v1 (digest SHA-256:${ref})
`, substitutionsV1))

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPONENTB_V1))
		})

		It("signs transport archive with --lookup option", func() {
			prepareEnv(env, ARCH2, ARCH)

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "--lookup", ARCH2, "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION, "--normalization", compdesc.JsonNormalisationV1)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${test}[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest SHA-256:${rb0}[genericBlobDigest/v1]
successfully signed github.com/mandelsoft/ref:v1 (digest SHA-256:${ref})
`, substitutionsV1))

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPONENTB_V1))
		})
	})

	Context("incomplete ctf", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
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

		It("sign version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(HaveOccurred())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Error: signing: github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: ocm reference "github.com/mandelsoft/test:v1" not found
`))
		})

		It("sign archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, ARCH)).To(HaveOccurred())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Error: signing: github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: ocm reference "github.com/mandelsoft/test:v1" not found
`))
		})
	})

	Context("incomplete component archive", func() {
		BeforeEach(func() {
			env.ComponentArchive(ARCH, accessio.FormatDirectory, COMPONENTB, VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("otherdata", "", "PlainText", metav1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "otherdata")
				})
				env.Reference("ref", COMPONENTA, VERSION)
			})
		})

		It("sign version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(HaveOccurred())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Error: signing: github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: ocm reference "github.com/mandelsoft/test:v1" not found
`))
		})

		It("sign archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, ARCH)).To(HaveOccurred())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Error: signing: github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: ocm reference "github.com/mandelsoft/test:v1" not found
`))
		})
	})

	Context("component archive", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENTA, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
				})
			})

			env.ComponentArchive(COMPARCH, accessio.FormatDirectory, COMPONENTB, VERSION, func() {
				env.Reference("ref", COMPONENTA, VERSION)
			})
		})

		It("signs comp arch with lookup", func() {
			buf := bytes.NewBuffer(nil)

			MustBeSuccessful(env.CatchOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--lookup", ARCH, "--repo", COMPARCH, "--normalization", compdesc.JsonNormalisationV1))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:5923de2b3b68e904eecb58eca91727926b36623623555025dc5a8700edfa9daa[jsonNormalisation/v1]
successfully signed github.com/mandelsoft/ref:v1 (digest SHA-256:3d1bf98adce06320809393473bed3aaaccf8696418bd1ef5b4d35fa632082d05)
`))
		})
	})

	It("keyless verification", func() {
		buf := bytes.NewBuffer(nil)

		// create Root CA
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=cerificate-authority", "root.priv")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair root.priv[root.cert]
`))
		Expect(env.FileExists("root.priv")).To(BeTrue())
		Expect(env.FileExists("root.cert")).To(BeTrue())

		// create CA used to create signing certificates
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=acme.org", "--ca-key", "root.priv", "--ca-cert", "root.cert", "ca.priv")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair ca.priv[ca.cert]
`))
		Expect(env.FileExists("ca.priv")).To(BeTrue())
		Expect(env.FileExists("ca.cert")).To(BeTrue())

		// create signing vcertificate from CA
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "CN=mandelsoft", "C=DE", "--ca-key", "ca.priv", "--ca-cert", "ca.cert", "--root-certs", "root.cert", "key.priv")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair key.priv[key.cert]
`))
		Expect(env.FileExists("key.priv")).To(BeTrue())
		Expect(env.FileExists("key.cert")).To(BeTrue())

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENTA, VERSION, func() {
				env.Provider("mandelsoft")
			})
		})

		// sigh component with certificate
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute("sign", "component", ARCH, "-K", "key.priv", "-k", "key.cert", "--ca-cert", "root.cert", "-s", "mandelsoft", "-I", "CN=mandelsoft", "--normalization", compdesc.JsonNormalisationV1)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
successfully signed github.com/mandelsoft/test:v1 (digest SHA-256:5ed8bb27309c3c2fff43f3b0f3ebb56a5737ad6db4bc8ace73c5455cb86faf54)
`))
		// verify component without key
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute("verify", "component", ARCH, "--ca-cert", "root.cert", "-I", "CN=mandelsoft")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
no public key found for signature "mandelsoft" -> extract key from signature
successfully verified github.com/mandelsoft/test:v1 (digest SHA-256:5ed8bb27309c3c2fff43f3b0f3ebb56a5737ad6db4bc8ace73c5455cb86faf54)
`))

		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
		defer Close(cv, "cv")

		Expect(len(cv.GetDescriptor().Signatures)).To(Equal(1))

		sig := cv.GetDescriptor().Signatures[0].Signature

		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(signutils.MediaTypePEM))

		_, algo, certs := Must3(signutils.GetSignatureFromPem([]byte(sig.Value)))
		Expect(len(certs)).To(Equal(3))
		Expect(algo).To(Equal(rsa.Algorithm))
	})

	Context("verified store", func() {
		BeforeEach(func() {
			FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIManifest1(env.Builder)
				OCIManifest2(env.Builder)
			})
		})

		DescribeTable("signs transport archive", func(substitutions Substitutions, normAlgo string) {
			prepareEnv(env, ARCH, ARCH)

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "--verified", VERIFIED_FILE, "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION, "--normalization", normAlgo)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${test}[${normAlgo}]
  resource 0:  "name"="otherdata": digest SHA-256:${rb0}[genericBlobDigest/v1]
successfully signed github.com/mandelsoft/ref:v1 (digest SHA-256:${ref})
`, substitutions))

			Expect(Must(env.FileExists(VERIFIED_FILE))).To(BeTrue())

			store := Must(signing.NewVerifiedStore(VERIFIED_FILE, env.FileSystem()))

			CheckStore(store, common.NewNameVersion(COMPONENTA, VERSION))
			CheckStore(store, common.NewNameVersion(COMPONENTB, VERSION))
		},
			Entry("v1", substitutionsV1, compdesc.JsonNormalisationV1),
			Entry("v2", substitutionsV2, compdesc.JsonNormalisationV2),
			Entry("v3", substitutionsV3, compdesc.JsonNormalisationV3),
		)
	})
})

func CheckStore(store signing.VerifiedStore, ve common.VersionedElement) {
	e := store.Get(ve)
	ExpectWithOffset(1, e).NotTo(BeNil())
	ExpectWithOffset(1, common.VersionedElementKey(e)).To(Equal(common.VersionedElementKey(ve)))
}

func prepareEnv(env *TestEnv, componentAArchive, componentBArchive string) {
	env.OCMCommonTransport(componentAArchive, accessio.FormatDirectory, func() {
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
	})

	if componentBArchive != "" {
		env.OCMCommonTransport(componentBArchive, accessio.FormatDirectory, func() {
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
	}
}
