package signing_test

import (
	"crypto/x509/pkix"
	"fmt"
	"time"

	. "github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/compositionmodeattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/resolvers"
	. "ocm.software/ocm/api/ocm/testhelper"
	. "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/ocm/tools/signing/signingtest"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/tech/signing/hasher/sha512"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
)

var DefaultContext = ocm.New()

const (
	ARCH       = "/tmp/ctf"
	PROVIDER   = "mandelsoft"
	VERSION    = "v1"
	COMPONENTA = "github.com/mandelsoft/test"
	COMPONENTB = "github.com/mandelsoft/ref"
	COMPONENTC = "github.com/mandelsoft/ref2"
	COMPONENTD = "github.com/mandelsoft/top"
	OUT        = "/tmp/res"
	OCIPATH    = "/tmp/oci"
	OCIHOST    = "alias"
)

const (
	SIGNATURE  = "test"
	SIGNATURE2 = "second"
	SIGN_ALGO  = rsa.Algorithm
)

var _ = Describe("access method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(signingtest.ModifiableTestData())
		env.RSAKeyPair(SIGNATURE, SIGNATURE2)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	/* TODO: add complex example from component cli
	Context("compatibility", func() {
		It("verifies older hash types (sha256[digest type] instead of SHA-256(crypto type))", func() {
			session := datacontext.NewSession()
			defer session.Close()

			env.ReadRSAKeyPair(SIGNATURE, "/testdata/compat")
			cv, err := comparch.Open(env.OCMContext(), accessobj.ACC_READONLY, "/testdata/compat/component-archive", 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			opts := NewOptions(
				VerifySignature(SIGNATURE),
				VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal("b06e4c1a68274b876661f9fbf1f100526d289745f6ee847bfef702007b5b14cf"))
			Expect(dig.HashAlgorithm).To(Equal(sha256.Algorithm))
		})
		It("resigns with older hash types", func() {
			session := datacontext.NewSession()
			defer session.Close()

			env.ReadRSAKeyPair(SIGNATURE, "/testdata/compat")
			cv, err := comparch.Open(env.OCMContext(), accessobj.ACC_READONLY, "/testdata/compat/component-archive", 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			opts := NewOptions(
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				VerifySignature(SIGNATURE),
				VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal("b06e4c1a68274b876661f9fbf1f100526d289745f6ee847bfef702007b5b14cf"))
			Expect(dig.HashAlgorithm).To(Equal(sha256.Algorithm))
		})
	})
	*/

	Context("special cases", func() {
		DescribeTable("handles none access", func(mode string) {
			env.ModificationOptions(ocm.SkipDigest())
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENTA, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						TestDataResource(env)
						env.Resource("value", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								none.New(),
							)
						})
					})
				})
			})

			session := datacontext.NewSession()
			defer session.Close()

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
			archcloser := session.AddCloser(src)
			resolver := resolvers.NewCompoundResolver(src)

			cv := Must(resolver.LookupComponentVersion(COMPONENTA, VERSION))
			closer := session.AddCloser(cv)

			digest := "123d48879559d16965a54eba9a3e845709770f4f0be984ec8db2f507aa78f338"

			pr, buf := common.NewBufferedPrinter()
			// key taken from signing attr
			dig := Must(SignComponentVersion(cv, SIGNATURE, SignerByAlgo(SIGN_ALGO), Resolver(resolver), DigestMode(mode), Printer(pr)))
			Expect(closer.Close()).To(Succeed())
			Expect(archcloser.Close()).To(Succeed())
			Expect(dig.Value).To(StringEqualWithContext(digest))

			src = Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			session.AddCloser(src)
			cv = Must(src.LookupComponentVersion(COMPONENTA, VERSION))
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digest))

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
  resource 0:  "name"="testdata": digest SHA-256:${D_TESTDATA}[genericBlobDigest/v1]
`, Digests))

			CheckResourceDigests(cv.GetDescriptor(), map[string]*metav1.DigestSpec{
				"testdata": DS_TESTDATA,
			})
			////////

			dig = Must(VerifyComponentVersion(cv, SIGNATURE, Resolver(resolver), Printer(pr)))
			Expect(dig.Value).To(Equal(digest))
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)
	})

	Context("valid", func() {
		digestA := "01de99400030e8336020059a435cea4e7fe8f21aad4faf619da882134b85569d"
		digestB := "5f416ec59629d6af91287e2ba13c6360339b6a0acf624af2abd2a810ce4aefce"

		localDigests := Substitutions{
			"D_COMPA": digestA,
			"D_COMPB": digestB,
		}
		BeforeEach(func() {
			FakeOCIRepo(env, OCIPATH, OCIHOST)

			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIManifest1(env)
				OCIManifest2(env)
			})

			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENTA, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						TestDataResource(env)
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
						OtherDataResource(env)
						env.Reference("ref", COMPONENTA, VERSION)
					})
				})
			})
		})

		DescribeTable("sign flat version", func(mode string) {
			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			archcloser := session.AddCloser(src)
			resolver := resolvers.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			closer := session.AddCloser(cv)

			opts := NewOptions(
				Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			pr, buf := common.NewBufferedPrinter()
			dig, err := Apply(pr, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(closer.Close()).To(Succeed())
			Expect(archcloser.Close()).To(Succeed())
			Expect(dig.Value).To(StringEqualWithContext(digestA))

			src, err = ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err = src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digestA))

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
  resource 0:  "name"="testdata": digest SHA-256:${D_TESTDATA}[genericBlobDigest/v1]
  resource 1:  "name"="value": digest SHA-256:${D_OCIMANIFEST1}[ociArtifactDigest/v1]
  resource 2:  "name"="ref": digest SHA-256:${D_OCIMANIFEST2}[ociArtifactDigest/v1]
`, Digests, OCIDigests))
			////////

			opts = NewOptions(
				DigestMode(mode),
				VerifySignature(SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			dig, err = Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal(digestA))

			CheckResourceDigests(cv.GetDescriptor(), map[string]*metav1.DigestSpec{
				"testdata": DS_TESTDATA,
				"value":    DS_OCIMANIFEST1,
				"ref":      DS_OCIMANIFEST2,
			})

			cv.GetDescriptor().Resources[0].Digest.Value = "010ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50" // some wrong value
			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(StringEqualWithContext("github.com/mandelsoft/test:v1: calculated resource digest (SHA-256:" + D_TESTDATA + "[genericBlobDigest/v1]) mismatches existing digest (SHA-256:010ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]) for testdata:v1 (Local blob sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[])"))
			// Reset to original to avoid write back in readonly mode
			cv.GetDescriptor().Resources[0].Digest.Value = D_TESTDATA

			cv.GetDescriptor().Signatures[0].Digest.Value = "0ae7ab0c1578d1292922b2a3884833c380a57df2cc7dfab7213ee051b092edc3"
			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(StringEqualTrimmedWithContext("github.com/mandelsoft/test:v1: signature digest (0ae7ab0c1578d1292922b2a3884833c380a57df2cc7dfab7213ee051b092edc3) does not match found digest (" + digestA + ")"))
			// Reset to original to avoid write back in readonly mode
			cv.GetDescriptor().Signatures[0].Digest.Value = digestA
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)

		DescribeTable("sign flat version with generic verification", func(mode string) {
			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			archcloser := session.AddCloser(src)
			resolver := resolvers.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			closer := session.AddCloser(cv)

			opts := NewOptions(
				DigestMode(mode),
				Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())
			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			closer.Close()
			archcloser.Close()
			Expect(dig.Value).To(StringEqualWithContext(digestA))

			src, err = ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err = src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digestA))

			////////

			opts = NewOptions(
				VerifySignature(),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			dig, err = Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal(digestA))
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)

		DescribeTable("sign deep version", func(mode string) {
			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			archcloser := session.AddCloser(src)
			resolver := resolvers.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			closer := session.AddCloser(cv)

			opts := NewOptions(
				DigestMode(mode),
				Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			pr, buf := common.NewBufferedPrinter()
			dig, err := Apply(pr, nil, cv, opts)
			Expect(err).To(Succeed())
			closer.Close()
			archcloser.Close()
			Expect(dig.Value).To(StringEqualWithContext(digestB))

			src, err = ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err = src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digestB))

			if mode == DIGESTMODE_TOP {
				Expect(cv.GetDescriptor().NestedDigests.String()).To(StringEqualTrimmedWithContext(`
github.com/mandelsoft/test:v1: SHA-256:${D_COMPA}[jsonNormalisation/v1]
  testdata:v1[]: SHA-256:${D_TESTDATA}[genericBlobDigest/v1]
  value:v1[]: SHA-256:${D_OCIMANIFEST1}[ociArtifactDigest/v1]
  ref:v1[]: SHA-256:${D_OCIMANIFEST2}[ociArtifactDigest/v1]
`, localDigests, Digests, OCIDigests))
			} else {
				Expect(cv.GetDescriptor().NestedDigests).To(BeNil())
			}

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="testdata": digest SHA-256:${D_TESTDATA}[genericBlobDigest/v1]
    resource 1:  "name"="value": digest SHA-256:${D_OCIMANIFEST1}[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest SHA-256:${D_OCIMANIFEST2}[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest SHA-256:${D_OTHERDATA}[genericBlobDigest/v1]
`, localDigests, Digests, OCIDigests))
			////////

			opts = NewOptions(
				VerifySignature(SIGNATURE),
				Resolver(src),
				VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			dig, err = Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal(digestB))
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)

		DescribeTable("fails generic verification", func(mode string) {
			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			resolver := resolvers.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)

			opts := NewOptions(
				DigestMode(mode),
				VerifySignature(),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(StringEqualWithContext("github.com/mandelsoft/test:v1: failed to determine signature info: no signature found"))
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)
	})

	Context("invalid", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENTB, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						OtherDataResource(env)
						env.Reference("ref", COMPONENTA, VERSION)
					})
				})
			})
		})

		DescribeTable("fails signing version with unknown ref", func(mode string) {
			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)

			opts := NewOptions(
				DigestMode(mode),
				Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(src),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())

			cv, err := src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)

			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(StringEqualWithContext("github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: component version \"github.com/mandelsoft/test:v1\" not found: oci artifact \"v1\" not found in component-descriptors/github.com/mandelsoft/test"))
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)
	})

	Context("legacy rhombus", func() {
		D_DATAA := "8a835d52867572bdaf7da7fb35ee59ad45c3db2dacdeeca62178edd5d07ef08c"
		D_DATAA512 := "47e63fa783ec370d83b84ce3de37a3de3fdd5cdc64d4fb21a530dce00fa1011e8dbc85f7509694a5875bf82e710ce00ac6bcd8e716741a7fc4c51a181b741920"
		D_DATAB := "5f103fcedc97b81bfc1841447d164781ed0f6244ce20b26d7a8a7d5880156c33"
		D_DATAB512 := "a9469fc2e9787c8496cf1526508ae86d4e855715ef6b8f7031bdc55759683762f1c330b94a4516dff23e32f19fb170cbcb53015f1ffc0d77624ee5c9a288a030"
		D_DATAC := "90e06e32c46338db42d78d49fee035063d4b10e83cfbf0d1831e14527245da12"
		D_DATAD := "5a5c3f681c2af10d682926a635a1dc9dfe7087d4fa3daf329bf0acad540911a9"

		DS_DATAA := TextResourceDigestSpec(D_DATAA)
		DS_DATAB := TextResourceDigestSpec(D_DATAB)
		DS_DATAC := TextResourceDigestSpec(D_DATAC)
		DS_DATAD := TextResourceDigestSpec(D_DATAD)

		D_COMPA := "bdb62ce8299f10e230b91bc9a9bfdbf2d33147f205fcf736d802c7e1cec7b5e8"
		D_COMPA512 := "7aa760f27b494814e56c44413afd7bc9d932df28918d63bea222be4bd2b6abd921225cca2140d6eb549418a75b8db2a32be1852012d77474657505f0ea57b34d"
		// D_COMPA_HASHED := "0bf5d019bab058a392b6bcb2ae50c93a02f623da0a439b1bbbfd4b1f795fbd3aafe271e3b757fad06e9118f74b18c2b83c7443f86e0c04c4539196bad79c6380"
		D_COMPB := "d1def1b60cc8b241451b0e3fccb705a9d99db188b72ec4548519017921700857"
		D_COMPB512 := "08366761127c791e550d2082e34e68c8836739c68f018f969a46a17a6c13b529390303335ee0ae3cd938af9e0f31665427a1b45360622d864a5dbe053917a75d"
		// D_COMPBR := "e47deeca35bc34116770a50a88954a0b028eb4e236d089b84e419c6d7ce15d97"
		D_COMPC := "b376a7b440c0b1e506e54a790966119a8e229cf9226980b84c628d77ef06fc58"
		D_COMPD := "64674d3e2843d36c603f44477e4cd66ee85fe1a91227bbcd271202429024ed61"

		localDigests := Substitutions{
			"D_DATAA":    D_DATAA,
			"D_DATAB":    D_DATAB,
			"D_DATAB512": D_DATAB512,
			"D_DATAC":    D_DATAC,
			"D_DATAD":    D_DATAD,

			"D_COMPA": D_COMPA,
			"D_COMPB": D_COMPB,
			// "D_COMPBR":   D_COMPB,
			"D_COMPC":    D_COMPC,
			"D_COMPD":    D_COMPD,
			"D_COMPB512": D_COMPB512,
		}

		_, _, _, _ = DS_DATAA, DS_DATAB, DS_DATAC, DS_DATAD

		setup := func(opts ...ocm.ModificationOption) {
			env.ModificationOptions(opts...)
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENTA, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("data_a", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata A")
						})
					})
				})
				env.Component(COMPONENTB, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("data_b", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata B")
						})
						env.Reference("ref", COMPONENTA, VERSION)
					})
				})
				env.Component(COMPONENTC, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("data_c", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata C")
						})
						env.Reference("ref", COMPONENTA, VERSION)
					})
				})
				env.Component(COMPONENTD, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("data_d", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata D")
						})
						env.Reference("refb", COMPONENTB, VERSION)
						env.Reference("refc", COMPONENTC, VERSION)
					})
				})
			})
		}

		DescribeTable("hashes unsigned", func(mode bool, c EntryCheck, mopts ...ocm.ModificationOption) {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			{
				compositionmodeattr.Set(env.OCMContext(), mode)
				setup(mopts...)
				arch := finalizer.Nested()
				src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
				Expect(err).To(Succeed())
				arch.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				log := HashComponent(resolver, COMPONENTD, D_COMPD, DigestMode(c.Mode()))

				Expect(log).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/top:v1"[github.com/mandelsoft/top:v1]...
  no digest found for "github.com/mandelsoft/ref:v1"
  applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/top:v1]...
    no digest found for "github.com/mandelsoft/test:v1"
    applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/top:v1]...
      resource 0:  "name"="data_a": digest SHA-256:${D_DATAA}[genericBlobDigest/v1]
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_b": digest SHA-256:${D_DATAB}[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/ref:v1: digest SHA-256:${D_COMPB}[jsonNormalisation/v1]
  no digest found for "github.com/mandelsoft/ref2:v1"
  applying to version "github.com/mandelsoft/ref2:v1"[github.com/mandelsoft/top:v1]...
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_c": digest SHA-256:${D_DATAC}[genericBlobDigest/v1]
  reference 1:  github.com/mandelsoft/ref2:v1: digest SHA-256:${D_COMPC}[jsonNormalisation/v1]
  resource 0:  "name"="data_d": digest SHA-256:${D_DATAD}[genericBlobDigest/v1]
`, localDigests))
				MustBeSuccessful(arch.Finalize())
			}

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			finalizer.Close(src)
			cv, err := src.LookupComponentVersion(COMPONENTD, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cv)
			Expect(len(cv.GetDescriptor().Signatures)).To(Equal(0))

			c.Check1CheckD(cv, localDigests)

			Expect(cv.GetDescriptor().Resources[0].Digest).NotTo(BeNil())
			Expect(cv.GetDescriptor().Resources[0].Digest.String()).To(Equal("SHA-256:" + D_DATAD + "[genericBlobDigest/v1]"))

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			sub := finalizer.Nested()
			sub.Close(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			c.Check1CheckA(cva, DS_DATAA, mopts...)

			////////

			VerifyHashes(src, COMPONENTD, D_COMPD)

			c.Check1Corrupt(cva, sub, cv)

			opts := NewOptions(
				Resolver(src),
				VerifyDigests(),
			)
			Expect(opts.Complete(env)).To(Succeed())
			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("github.com/mandelsoft/top:v1: failed applying to component reference refb[github.com/mandelsoft/ref:v1]: github.com/mandelsoft/top:v1->github.com/mandelsoft/ref:v1: failed applying to component reference ref[github.com/mandelsoft/test:v1]: github.com/mandelsoft/top:v1->github.com/mandelsoft/ref:v1->github.com/mandelsoft/test:v1: calculated resource digest (SHA-256:" + D_DATAA + "[genericBlobDigest/v1]) mismatches existing digest (SHA-256:" + wrongDigest + "[genericBlobDigest/v1]) for data_a:v1 (Local blob sha256:" + D_DATAA + "[])"))
		},
			Entry(DIGESTMODE_TOP, false, &EntryTop{}),
			Entry(DIGESTMODE_LOCAL, false, &EntryLocal{}),

			Entry("legacy "+DIGESTMODE_TOP, false, &EntryTop{}, ocm.SkipDigest()),
			Entry("legacy "+DIGESTMODE_LOCAL, false, &EntryLocal{}, ocm.SkipDigest()),

			Entry(DIGESTMODE_TOP+" with composition mode", true, &EntryTop{}),
			Entry(DIGESTMODE_LOCAL+" with composition mode", true, &EntryLocal{}),

			Entry("legacy "+DIGESTMODE_TOP+" with composition mode", true, &EntryTop{}, ocm.SkipDigest()),
			Entry("legacy "+DIGESTMODE_LOCAL+" with composition mode", true, &EntryLocal{}, ocm.SkipDigest()),
		)

		DescribeTable("signs unsigned", func(c EntryCheck, mopts ...ocm.ModificationOption) {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			{
				setup(mopts...)
				arch := finalizer.Nested()
				src, err := ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env)
				Expect(err).To(Succeed())
				arch.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				log := SignComponent(resolver, SIGNATURE, COMPONENTD, D_COMPD, DigestMode(c.Mode()))

				Expect(log).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/top:v1"[github.com/mandelsoft/top:v1]...
  no digest found for "github.com/mandelsoft/ref:v1"
  applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/top:v1]...
    no digest found for "github.com/mandelsoft/test:v1"
    applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/top:v1]...
      resource 0:  "name"="data_a": digest SHA-256:${D_DATAA}[genericBlobDigest/v1]
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_b": digest SHA-256:${D_DATAB}[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/ref:v1: digest SHA-256:${D_COMPB}[jsonNormalisation/v1]
  no digest found for "github.com/mandelsoft/ref2:v1"
  applying to version "github.com/mandelsoft/ref2:v1"[github.com/mandelsoft/top:v1]...
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_c": digest SHA-256:${D_DATAC}[genericBlobDigest/v1]
  reference 1:  github.com/mandelsoft/ref2:v1: digest SHA-256:${D_COMPC}[jsonNormalisation/v1]
  resource 0:  "name"="data_d": digest SHA-256:${D_DATAD}[genericBlobDigest/v1]
`, localDigests))
				MustBeSuccessful(arch.Finalize())
			}
			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			finalizer.Close(src)
			cv, err := src.LookupComponentVersion(COMPONENTD, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPD))

			c.Check1CheckD(cv, localDigests)
			c.Check2Ref(cv, "refb", D_COMPB)
			c.Check2Ref(cv, "refc", D_COMPC)

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			c.Check1CheckA(cva, DS_DATAA, mopts...)
			////////

			cvb := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
			finalizer.Close(cvb)
			Expect(len(cvb.GetDescriptor().Signatures)).To(Equal(0))
			c.Check2Ref(cvb, "ref", D_COMPA)

			cvc := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
			finalizer.Close(cvc)
			Expect(len(cvb.GetDescriptor().Signatures)).To(Equal(0))
			c.Check2Ref(cvb, "ref", D_COMPA)

			VerifyComponent(src, SIGNATURE, COMPONENTD, D_COMPD)
		},
			Entry(DIGESTMODE_TOP, &EntryTop{}),
			Entry(DIGESTMODE_LOCAL, &EntryLocal{}),

			Entry("legacy "+DIGESTMODE_TOP, &EntryTop{}, ocm.SkipDigest()),
			Entry("legacy "+DIGESTMODE_LOCAL, &EntryLocal{}, ocm.SkipDigest()),
		)

		// D_COMPD_LEGACY := "342d30317bee13ec30d815122f23b19d9ee54a15ff8be1ec550c8072d5a6dba6"
		// D_COMPB_HASHED := "af8a8324b7848fc5887c63e632402df99c729669889b4d2ae7efceb9f1c2341b81d8a18b82e994564d854422a544c3dffc7d64d8389c90ab7fad19a50bb75e31"
		D_COMPB_HASHED := "6ef2fa650b73302f2f23543adf4588e18ec419c5604eab43dcbb7d4ef12a7e6ad0f5d872d34a9839d428861e22770973e0ca7316891f8b246cb0942d4fede3fc"
		// DigestDFor512 := "64674d3e2843d36c603f44477e4cd66ee85fe1a91227bbcd271202429024ed61"
		// DigestBFor512 := "af8a8324b7848fc5887c63e632402df99c729669889b4d2ae7efceb9f1c2341b81d8a18b82e994564d854422a544c3dffc7d64d8389c90ab7fad19a50bb75e31"

		DescribeTable("signs and rehashes presigned in top mode", (func(subst Substitutions, mopts ...ocm.ModificationOption) {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			setup(mopts...)
			{
				arch := finalizer.Nested()
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				arch.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				log := SignComponent(resolver, SIGNATURE, COMPONENTB, subst["D_COMPB_X"], DigestMode(DIGESTMODE_TOP), HashByAlgo(sha512.Algorithm))
				Expect(log).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref:v1]...
  no digest found for "github.com/mandelsoft/test:v1"
  applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref:v1]...
    resource 0:  "name"="data_a": digest SHA-512:${D_DATAA_X}[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest SHA-512:${D_COMPA_X}[jsonNormalisation/v1]
  resource 0:  "name"="data_b": digest ${HASH}:${D_DATAB_X}[genericBlobDigest/v1]

`, MergeSubst(localDigests, subst)))
				VerifyComponent(resolver, SIGNATURE, COMPONENTB, subst["D_COMPB_X"])
				log = SignComponent(resolver, SIGNATURE, COMPONENTD, subst["D_COMPD_X"], DigestMode(DIGESTMODE_TOP))

				Expect(log).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/top:v1"[github.com/mandelsoft/top:v1]...
  no digest found for "github.com/mandelsoft/ref:v1"
  applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/top:v1]...
    no digest found for "github.com/mandelsoft/test:v1"
    applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/top:v1]...
      resource 0:  "name"="data_a": digest SHA-256:${D_DATAA}[genericBlobDigest/v1]
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_b": digest SHA-256:${D_DATAB}[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/ref:v1: digest SHA-256:${D_COMPB}[jsonNormalisation/v1]
  no digest found for "github.com/mandelsoft/ref2:v1"
  applying to version "github.com/mandelsoft/ref2:v1"[github.com/mandelsoft/top:v1]...
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_c": digest SHA-256:${D_DATAC}[genericBlobDigest/v1]
  reference 1:  github.com/mandelsoft/ref2:v1: digest SHA-256:${D_COMPC}[jsonNormalisation/v1]
  resource 0:  "name"="data_d": digest SHA-256:${D_DATAD}[genericBlobDigest/v1]
`, MergeSubst(localDigests, subst)))
				Defer(arch.Finalize)
			}

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			finalizer.Close(src)
			cv := Must(src.LookupComponentVersion(COMPONENTD, VERSION))
			finalizer.Close(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(subst["D_COMPD_X"]))

			Expect(cv.GetDescriptor().NestedDigests).NotTo(BeNil())
			Expect(cv.GetDescriptor().NestedDigests.String()).To(StringEqualTrimmedWithContext(`
github.com/mandelsoft/ref:v1: SHA-256:${D_COMPB}[jsonNormalisation/v1]
  data_b:v1[]: SHA-256:${D_DATAB}[genericBlobDigest/v1]
github.com/mandelsoft/ref2:v1: SHA-256:${D_COMPC}[jsonNormalisation/v1]
  data_c:v1[]: SHA-256:${D_DATAC}[genericBlobDigest/v1]
github.com/mandelsoft/test:v1: SHA-256:${D_COMPA}[jsonNormalisation/v1]
  data_a:v1[]: SHA-256:${D_DATAA}[genericBlobDigest/v1]
`, MergeSubst(localDigests, subst)))

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			////////

			VerifyComponent(src, SIGNATURE, COMPONENTD, subst["D_COMPD_X"])
		}),
			Entry("legacy", Substitutions{
				"HASH":      "SHA-512",
				"D_DATAA_X": D_DATAA512,
				"D_DATAB_X": D_DATAB512,
				"D_DATAB":   D_DATAB,
				"D_COMPA_X": D_COMPA512,
				"D_COMPB_X": D_COMPB512,
				"D_COMPD_X": D_COMPD,
				"D_COMPB":   D_COMPB,
			}, ocm.SkipDigest()),
			Entry("hashed", Substitutions{
				"HASH":      "SHA-256",
				"D_DATAA_X": D_DATAA512,
				"D_DATAB_X": D_DATAB,
				"D_COMPA_X": D_COMPA512,
				"D_COMPB_X": D_COMPB_HASHED,
				"D_COMPD_X": D_COMPD,
			}),
		)

		DescribeTable("verifies after sub level signing", func(subst Substitutions, mopts ...ocm.ModificationOption) {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			setup(mopts...)

			digestD := D_COMPD
			{
				arch := finalizer.Nested()
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				arch.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				fmt.Printf("SIGN D\n")
				log := SignComponent(resolver, SIGNATURE, COMPONENTD, digestD, DigestMode(DIGESTMODE_TOP))

				Expect(log).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/top:v1"[github.com/mandelsoft/top:v1]...
  no digest found for "github.com/mandelsoft/ref:v1"
  applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/top:v1]...
    no digest found for "github.com/mandelsoft/test:v1"
    applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/top:v1]...
      resource 0:  "name"="data_a": digest SHA-256:${D_DATAA}[genericBlobDigest/v1]
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_b": digest SHA-256:${D_DATAB}[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/ref:v1: digest SHA-256:${D_COMPB}[jsonNormalisation/v1]
  no digest found for "github.com/mandelsoft/ref2:v1"
  applying to version "github.com/mandelsoft/ref2:v1"[github.com/mandelsoft/top:v1]...
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:${D_COMPA}[jsonNormalisation/v1]
    resource 0:  "name"="data_c": digest SHA-256:${D_DATAC}[genericBlobDigest/v1]
  reference 1:  github.com/mandelsoft/ref2:v1: digest SHA-256:${D_COMPC}[jsonNormalisation/v1]
  resource 0:  "name"="data_d": digest SHA-256:${D_DATAD}[genericBlobDigest/v1]
`, MergeSubst(localDigests, subst)))

				fmt.Printf("SIGN B\n")
				SignComponent(resolver, SIGNATURE, COMPONENTB, subst["D_COMPB_X"], HashByAlgo(sha512.Algorithm), DigestMode(DIGESTMODE_TOP))
				fmt.Printf("VERIFY B\n")
				VerifyComponent(resolver, SIGNATURE, COMPONENTB, subst["D_COMPB_X"])

				Defer(arch.Finalize)
			}

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			finalizer.Close(src)
			cv, err := src.LookupComponentVersion(COMPONENTD, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digestD))

			Expect(cv.GetDescriptor().NestedDigests.String()).To(StringEqualTrimmedWithContext(`
github.com/mandelsoft/ref:v1: SHA-256:${D_COMPB}[jsonNormalisation/v1]
  data_b:v1[]: SHA-256:${D_DATAB}[genericBlobDigest/v1]
github.com/mandelsoft/ref2:v1: SHA-256:${D_COMPC}[jsonNormalisation/v1]
  data_c:v1[]: SHA-256:${D_DATAC}[genericBlobDigest/v1]
github.com/mandelsoft/test:v1: SHA-256:${D_COMPA}[jsonNormalisation/v1]
  data_a:v1[]: SHA-256:${D_DATAA}[genericBlobDigest/v1]
`, MergeSubst(localDigests, subst)))

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			////////
			fmt.Printf("VERIFY D\n")
			VerifyComponent(src, SIGNATURE, COMPONENTD, digestD)
		},
			Entry("legacy", Substitutions{
				"D_COMPB_X": D_COMPB512,
			}, ocm.SkipDigest()),
			Entry("hashed", Substitutions{
				"D_COMPB_X": D_COMPB_HASHED,
			}),
		)

		DescribeTable("fixes digest mode", func(subst Substitutions, mopts ...ocm.ModificationOption) {
			setup(mopts...)

			var finalizer Finalizer
			defer Check(finalizer.Finalize)

			{ // sign with mode local
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				finalizer.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				fmt.Printf("SIGN B\n")
				_ = SignComponent(resolver, SIGNATURE, COMPONENTB, D_COMPB, DigestMode(DIGESTMODE_LOCAL))
				VerifyComponent(src, SIGNATURE, COMPONENTB, D_COMPB)
				Check(finalizer.Finalize)
			}
			{ // check mode
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
				finalizer.Close(src)
				cv := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
				finalizer.Close(cv)
				Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPB))
				Expect(cv.GetDescriptor().NestedDigests).To(BeNil())
				Expect(cv.GetDescriptor().References[0].Digest).NotTo(BeNil())
				Expect(GetDigestMode(cv.GetDescriptor())).To(Equal(DIGESTMODE_LOCAL))
				Check(finalizer.Finalize)
			}
			{ // try resign with mode top
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				finalizer.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				fmt.Printf("RESIGN B\n")
				_ = SignComponent(resolver, SIGNATURE, COMPONENTB, D_COMPB, DigestMode(DIGESTMODE_TOP), SignatureName(SIGNATURE2, true))
				VerifyComponent(src, SIGNATURE, COMPONENTB, D_COMPB, SignatureName(SIGNATURE2, true))
				Check(finalizer.Finalize)
			}
			{ // check mode
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
				finalizer.Close(src)
				cv := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
				finalizer.Close(cv)
				cd := cv.GetDescriptor()
				Expect(len(cd.Signatures)).To(Equal(2))
				Expect(cd.Signatures[0].Digest.Value).To(Equal(D_COMPB))
				Expect(cd.Signatures[0].Name).To(Equal(SIGNATURE))
				Expect(cd.Signatures[1].Digest.Value).To(Equal(D_COMPB))
				Expect(cd.Signatures[1].Name).To(Equal(SIGNATURE2))
				Expect(cv.GetDescriptor().NestedDigests).To(BeNil())
				Expect(cv.GetDescriptor().References[0].Digest).NotTo(BeNil())
				Expect(GetDigestMode(cd)).To(Equal(DIGESTMODE_LOCAL))
				Check(finalizer.Finalize)
			}
		},
			Entry("legacy", Substitutions{}, ocm.SkipDigest()),
			Entry("hashed", Substitutions{}),
		)

		DescribeTable("fixes digest mode in recursive signing", func(subst Substitutions, mopts ...ocm.ModificationOption) {
			var finalizer Finalizer
			defer Check(finalizer.Finalize)

			setup(mopts...)

			{ // sign with mode local
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				finalizer.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				fmt.Printf("SIGN B\n")
				_ = SignComponent(resolver, SIGNATURE, COMPONENTB, D_COMPB, DigestMode(DIGESTMODE_TOP), SignatureName(SIGNATURE2, true))
				VerifyComponent(src, SIGNATURE, COMPONENTB, D_COMPB, SignatureName(SIGNATURE2, true))
				Check(finalizer.Finalize)
			}
			{ // check mode
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
				finalizer.Close(src)
				cv := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
				finalizer.Close(cv)
				Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPB))
				Expect(cv.GetDescriptor().NestedDigests).NotTo(BeNil())
				Expect(cv.GetDescriptor().References[0].Digest).To(BeNil())
				Expect(GetDigestMode(cv.GetDescriptor())).To(Equal(DIGESTMODE_TOP))
				Check(finalizer.Finalize)
			}
			{ // resign recursively from top
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				finalizer.Close(src)

				resolver := resolvers.NewCompoundResolver(src)

				fmt.Printf("SIGN D\n")
				_ = SignComponent(resolver, SIGNATURE, COMPONENTD, D_COMPD, Recursive(), DigestMode(DIGESTMODE_LOCAL))
				VerifyComponent(src, SIGNATURE, COMPONENTD, D_COMPD)
				Check(finalizer.Finalize)
			}
			{ // check mode
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
				finalizer.Close(src)

				cv := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
				finalizer.Close(cv)
				cd := cv.GetDescriptor()
				Expect(len(cd.Signatures)).To(Equal(2))
				Expect(cd.Signatures[0].Digest.Value).To(Equal(D_COMPB))
				Expect(cd.Signatures[0].Name).To(Equal(SIGNATURE2))
				Expect(cd.Signatures[1].Digest.Value).To(Equal(D_COMPB))
				Expect(cd.Signatures[1].Name).To(Equal(SIGNATURE))
				Expect(cd.NestedDigests).NotTo(BeNil())
				Expect(cd.References[0].Digest).To(BeNil())
				Expect(GetDigestMode(cd)).To(Equal(DIGESTMODE_TOP))

				cv = Must(src.LookupComponentVersion(COMPONENTD, VERSION))
				finalizer.Close(cv)
				cd = cv.GetDescriptor()
				Expect(len(cd.Signatures)).To(Equal(1))
				Expect(cd.Signatures[0].Digest.Value).To(Equal(D_COMPD))
				Expect(cd.Signatures[0].Name).To(Equal(SIGNATURE))
				Expect(cv.GetDescriptor().NestedDigests).To(BeNil())
				Expect(cv.GetDescriptor().References[0].Digest).NotTo(BeNil())
				Expect(GetDigestMode(cd)).To(Equal(DIGESTMODE_LOCAL))

				Check(finalizer.Finalize)
			}
		},
			Entry("legacy", Substitutions{}, ocm.SkipDigest()),
			Entry("hashed", Substitutions{}),
		)
	})

	Context("ref hashes", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENTA, VERSION, func() {
					env.Provider(PROVIDER)
				})
				env.ComponentVersion(COMPONENTB, VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("refa", COMPONENTA, VERSION)
				})
				env.ComponentVersion(COMPONENTC, VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("refb", COMPONENTB, VERSION)
				})
			})
		})

		It("handles top level signature", func() {
			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
			defer Close(src, "ctf")

			resolver := resolvers.NewCompoundResolver(src)

			cv := Must(resolver.LookupComponentVersion(COMPONENTC, VERSION))
			defer cv.Close()

			opts := NewOptions(
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				VerifyDigests(),
			)
			MustBeSuccessful(opts.Complete(env))

			digestC := "1e81ac0fe69614e6fd73ab7a1c809dd31fcbcb810f0036be7a296d226e4bd64b"
			pr, buf := common.NewBufferedPrinter()
			dig := Must(Apply(pr, nil, cv, opts))
			Expect(dig.Value).To(StringEqualWithContext(digestC))

			Expect(cv.GetDescriptor().References[0].Digest.HashAlgorithm).To(Equal(sha256.Algorithm))

			cvb := Must(resolver.LookupComponentVersion(COMPONENTB, VERSION))
			defer Close(cvb)
			Expect(cvb.GetDescriptor().References[0].Digest).NotTo(BeNil())
			_ = buf
		})
	})

	Context("keyless verification", func() {
		ca, capriv := Must2(rsa.CreateRootCertificate(signutils.CommonName("ca-authority"), 10*time.Hour))
		intercert, interpem, interpriv := Must3(rsa.CreateSigningCertificate(signutils.CommonName("acme.org"), ca, ca, capriv, 5*time.Hour, true))
		certIssuer := &pkix.Name{
			CommonName:    PROVIDER,
			Country:       []string{"DE", "US"},
			Locality:      []string{"Walldorf d"},
			StreetAddress: []string{"x y"},
			PostalCode:    []string{"69169"},
			Province:      []string{"BW"},
		}
		cert, pemBytes, priv := Must3(rsa.CreateSigningCertificate(certIssuer, interpem, ca, interpriv, time.Hour))

		certs := Must(signutils.GetCertificateChain(pemBytes, false))
		Expect(len(certs)).To(Equal(3))

		ctx := ocm.DefaultContext()

		var cv ocm.ComponentVersionAccess

		BeforeEach(func() {
			cv = composition.NewComponentVersion(ctx, COMPONENTA, VERSION)
		})

		It("is consistent", func() {
			MustBeSuccessful(signutils.VerifyCertificate(intercert, interpem, ca, nil))
			MustBeSuccessful(signutils.VerifyCertificate(cert, pemBytes, ca, nil))
		})

		It("signs with certificate and default issuer", func() {
			digest := "9cf14695c864411cad03071a8766e6769bb00373bdd8c65887e4644cc285dc78"
			res := resolvers.NewDedicatedResolver(cv)

			buf := SignComponent(res, PROVIDER, COMPONENTA, digest, PrivateKey(PROVIDER, priv), PublicKey(PROVIDER, pemBytes), RootCertificates(ca))
			Expect(buf).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
`))

			i := cv.GetDescriptor().GetSignatureIndex(PROVIDER)
			Expect(i).To(BeNumerically(">=", 0))
			sig := cv.GetDescriptor().Signatures[i].Signature
			Expect(sig.MediaType).To(Equal(signutils.MediaTypePEM))
			_, algo, chain := Must3(signutils.GetSignatureFromPem([]byte(sig.Value)))
			Expect(algo).To(Equal(rsa.Algorithm))
			Expect(len(chain)).To(Equal(3))

			VerifyComponent(res, PROVIDER, COMPONENTA, digest, RootCertificates(ca))
		})

		It("signs with certificate and explicit CN issuer", func() {
			digest := "9cf14695c864411cad03071a8766e6769bb00373bdd8c65887e4644cc285dc78"
			res := resolvers.NewDedicatedResolver(cv)

			buf := SignComponent(res, SIGNATURE, COMPONENTA, digest, PrivateKey(SIGNATURE, priv), PublicKey(SIGNATURE, pemBytes), RootCertificates(ca), Issuer(PROVIDER))
			Expect(buf).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
`))

			i := cv.GetDescriptor().GetSignatureIndex(SIGNATURE)
			Expect(i).To(BeNumerically(">=", 0))
			sig := cv.GetDescriptor().Signatures[i].Signature
			Expect(sig.MediaType).To(Equal(signutils.MediaTypePEM))
			_, algo, chain := Must3(signutils.GetSignatureFromPem([]byte(sig.Value)))
			Expect(algo).To(Equal(rsa.Algorithm))
			Expect(len(chain)).To(Equal(3))

			VerifyComponent(res, SIGNATURE, COMPONENTA, digest, RootCertificates(ca), Issuer(PROVIDER))

			FailVerifyComponent(res, SIGNATURE, COMPONENTA, digest,
				`github.com/mandelsoft/test:v1: public key from signature: public key certificate: issuer mismatch in public key certificate: common name "mandelsoft" is invalid`,
				RootCertificates(ca))
		})

		It("signs with certificate and issuer", func() {
			digest := "9cf14695c864411cad03071a8766e6769bb00373bdd8c65887e4644cc285dc78"
			res := resolvers.NewDedicatedResolver(cv)
			issuer := &pkix.Name{
				CommonName: PROVIDER,
				Country:    []string{"DE"},
			}

			buf := SignComponent(res, SIGNATURE, COMPONENTA, digest, PrivateKey(SIGNATURE, priv), PublicKey(SIGNATURE, pemBytes), RootCertificates(ca), PKIXIssuer(*issuer))
			Expect(buf).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
`))

			i := cv.GetDescriptor().GetSignatureIndex(SIGNATURE)
			Expect(i).To(BeNumerically(">=", 0))
			sig := cv.GetDescriptor().Signatures[i].Signature
			Expect(sig.MediaType).To(Equal(signutils.MediaTypePEM))
			_, algo, chain := Must3(signutils.GetSignatureFromPem([]byte(sig.Value)))
			Expect(algo).To(Equal(rsa.Algorithm))
			Expect(len(chain)).To(Equal(3))
			dn := Must(signutils.ParseDN(sig.Issuer))
			Expect(dn).To(Equal(certIssuer))

			VerifyComponent(res, SIGNATURE, COMPONENTA, digest, RootCertificates(ca), PKIXIssuer(*issuer))

			issuer.Country = []string{"XX"}
			FailVerifyComponent(res, SIGNATURE, COMPONENTA, digest,
				`github.com/mandelsoft/test:v1: public key from signature: public key certificate: issuer mismatch in public key certificate: country "XX" not found`,
				RootCertificates(ca), PKIXIssuer(*issuer))
		})

		It("signs with certificate, issuer and tsa", func() {
			digest := "9cf14695c864411cad03071a8766e6769bb00373bdd8c65887e4644cc285dc78"
			res := resolvers.NewDedicatedResolver(cv)
			issuer := &pkix.Name{
				CommonName: "mandelsoft",
				Country:    []string{"DE"},
			}

			buf := SignComponent(res, SIGNATURE, COMPONENTA, digest, PrivateKey(SIGNATURE, priv), PublicKey(SIGNATURE, pemBytes), RootCertificates(ca), PKIXIssuer(*issuer), UseTSA())
			Expect(buf).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
`))

			i := cv.GetDescriptor().GetSignatureIndex(SIGNATURE)
			Expect(i).To(BeNumerically(">=", 0))
			sig := cv.GetDescriptor().Signatures[i]
			Expect(sig.Signature.MediaType).To(Equal(signutils.MediaTypePEM))
			_, algo, chain := Must3(signutils.GetSignatureFromPem([]byte(sig.Signature.Value)))
			Expect(algo).To(Equal(rsa.Algorithm))
			Expect(len(chain)).To(Equal(3))
			dn := Must(signutils.ParseDN(sig.Signature.Issuer))
			Expect(dn).To(Equal(certIssuer))

			Expect(sig.Timestamp).NotTo(BeNil())
			Expect(sig.Timestamp.Value).NotTo(Equal(""))
			Expect(sig.Timestamp.Time).NotTo(BeNil())
			Expect(time.Now().Sub(sig.Timestamp.Time.Time()).Minutes()).To(BeNumerically("<", 2))
			VerifyComponent(res, SIGNATURE, COMPONENTA, digest, RootCertificates(ca), PKIXIssuer(*issuer))

			issuer.Country = []string{"XX"}
			FailVerifyComponent(res, SIGNATURE, COMPONENTA, digest,
				`github.com/mandelsoft/test:v1: public key from signature: public key certificate: issuer mismatch in public key certificate: country "XX" not found`,
				RootCertificates(ca), PKIXIssuer(*issuer))
		})
	})

	Context("verified store", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENTA, VERSION, func() {
					env.Provider(PROVIDER)
				})
				env.ComponentVersion(COMPONENTB, VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("refa", COMPONENTA, VERSION)
				})
				env.ComponentVersion(COMPONENTC, VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("refb", COMPONENTB, VERSION)
				})
			})
		})

		It("remembers all indirectly signed component descriptors", func() {
			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
			defer Close(src, "ctf")

			resolver := resolvers.NewCompoundResolver(src)

			cv := Must(resolver.LookupComponentVersion(COMPONENTC, VERSION))
			defer Close(cv, "cv")

			store := NewLocalVerifiedStore()
			opts := NewOptions(
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				VerifyDigests(),
				UseVerifiedStore(store),
			)
			MustBeSuccessful(opts.Complete(env))

			pr, buf := common.NewBufferedPrinter()
			Must(Apply(pr, nil, cv, opts))

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/ref2:v1"[github.com/mandelsoft/ref2:v1]...
  no digest found for "github.com/mandelsoft/ref:v1"
  applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/ref2:v1]...
    no digest found for "github.com/mandelsoft/test:v1"
    applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/ref2:v1]...
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:5ed8bb27309c3c2fff43f3b0f3ebb56a5737ad6db4bc8ace73c5455cb86faf54[jsonNormalisation/v1]
  reference 0:  github.com/mandelsoft/ref:v1: digest SHA-256:e85e324ff16bafe26db235567d9232319c36f48ce995aa3f4957e55002207277[jsonNormalisation/v1]
`))

			CheckStore(store, cv)
			CheckStore(store, common.NewNameVersion(COMPONENTB, VERSION))
			CheckStore(store, common.NewNameVersion(COMPONENTA, VERSION))
		})
	})

	Context("handle extra identity", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENTA, VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("test", "v1", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "test data")
					})
					env.Resource("test", "v2", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "extended test data")
						env.ModificationOptions(ocm.AppendElement)
						env.ExtraIdentities()
					})
				})
			})
		})

		It("signs version with non-unique resource names", func() {
			session := datacontext.NewSession()
			defer session.Close()

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
			archcloser := session.AddCloser(src)

			cv := Must(src.LookupComponentVersion(COMPONENTA, VERSION))
			closer := session.AddCloser(cv)

			cd := cv.GetDescriptor()

			Expect(cd.Resources[0].GetIdentity(cv.GetDescriptor().Resources)).To(YAMLEqual(`
name: test
version: v1
`))
			Expect(cd.Resources[0].ExtraIdentity).To(YAMLEqual(`
version: v1
`))
			Expect(cd.Resources[1].GetIdentity(cv.GetDescriptor().Resources)).To(YAMLEqual(`
name: test
version: v2
`))
			Expect(cd.Resources[1].ExtraIdentity).To(YAMLEqual(`
version: v2
`))
			data := Must(compdesc.Encode(cd, compdesc.DefaultYAMLCodec))
			Expect(string(data)).To(YAMLEqual(`
  component:
    componentReferences: []
    name: github.com/mandelsoft/test
    provider: mandelsoft
    repositoryContexts: []
    resources:
    - access:
        localReference: sha256:916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9
        mediaType: text/plain
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: genericBlobDigest/v1
        value: 916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9
      name: test
      relation: external
      type: plainText
      version: v1
      extraIdentity:
        version: v1
    - access:
        localReference: sha256:920ce99fb13b43ca0408caee6e61f6335ea5156d79aa98e733e1ed2393e0f649
        mediaType: text/plain
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: genericBlobDigest/v1
        value: 920ce99fb13b43ca0408caee6e61f6335ea5156d79aa98e733e1ed2393e0f649
      name: test
      relation: external
      type: plainText
      version: v2
      extraIdentity:
        version: v2
    sources: []
    version: v1
  meta:
    schemaVersion: v2
`))

			digest := "70c1b7f5e2260a283e24788c81ea7f8f6e9a70a8544dbf62d6f3a27285f6b633"

			pr, buf := common.NewBufferedPrinter()
			// key taken from signing attr
			dig := Must(SignComponentVersion(cv, SIGNATURE, SignerByAlgo(SIGN_ALGO), Printer(pr)))
			Expect(closer.Close()).To(Succeed())
			Expect(archcloser.Close()).To(Succeed())
			Expect(dig.Value).To(StringEqualWithContext(digest))

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
  resource 0:  "name"="test","version"="v1": digest SHA-256:916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9[genericBlobDigest/v1]
  resource 1:  "name"="test","version"="v2": digest SHA-256:920ce99fb13b43ca0408caee6e61f6335ea5156d79aa98e733e1ed2393e0f649[genericBlobDigest/v1]
`))

			src = Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			session.AddCloser(src)
			cv = Must(src.LookupComponentVersion(COMPONENTA, VERSION))
			session.AddCloser(cv)

			cd = cv.GetDescriptor().Copy()
			Expect(len(cd.Signatures)).To(Equal(1))
			cd.Signatures = nil // for comparison
			data = Must(compdesc.Encode(cd, compdesc.DefaultYAMLCodec))

			Expect(string(data)).To(YAMLEqual(`
  component:
    componentReferences: []
    name: github.com/mandelsoft/test
    provider: mandelsoft
    repositoryContexts: []
    resources:
    - access:
        localReference: sha256:916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9
        mediaType: text/plain
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: genericBlobDigest/v1
        value: 916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9
      name: test
      relation: external
      type: plainText
      version: v1
      extraIdentity:
        version: v1
    - access:
        localReference: sha256:920ce99fb13b43ca0408caee6e61f6335ea5156d79aa98e733e1ed2393e0f649
        mediaType: text/plain
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: genericBlobDigest/v1
        value: 920ce99fb13b43ca0408caee6e61f6335ea5156d79aa98e733e1ed2393e0f649
      name: test
      relation: external
      type: plainText
      version: v2
      extraIdentity:
        version: v2
    sources: []
    version: v1
  meta:
    schemaVersion: v2
`))
		})
	})
})

func CheckStore(store VerifiedStore, ve common.VersionedElement) {
	e := store.Get(ve)
	ExpectWithOffset(1, e).NotTo(BeNil())
	ExpectWithOffset(1, common.VersionedElementKey(e)).To(Equal(common.VersionedElementKey(ve)))
}

func HashComponent(resolver ocm.ComponentVersionResolver, name string, digest string, other ...Option) string {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	Expect(err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		Resolver(resolver),
		Update(), VerifyDigests(),
	)
	opts.Eval(other...)
	ExpectWithOffset(1, opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

	pr, buf := common.NewBufferedPrinter()
	dig, err := Apply(pr, nil, cv, opts)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, dig.Value).To(StringEqualWithContext(digest))
	return buf.String()
}

func VerifyHashes(resolver ocm.ComponentVersionResolver, name string, digest string) {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	ExpectWithOffset(1, err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		Resolver(resolver),
		VerifyDigests(),
	)
	ExpectWithOffset(1, opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())
	dig, err := Apply(nil, nil, cv, opts)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, dig.Value).To(Equal(digest))
}

func SignComponent(resolver ocm.ComponentVersionResolver, signame, name string, digest string, other ...Option) string {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	Expect(err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		Sign(signingattr.Get(cv.GetContext()).GetSigner(SIGN_ALGO), signame),
		Resolver(resolver),
		VerifyDigests(),
	)
	opts.Eval(other...)
	ExpectWithOffset(1, opts.Complete(cv.GetContext())).To(Succeed())

	pr, buf := common.NewBufferedPrinter()
	dig, err := Apply(pr, nil, cv, opts)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, dig.Value).To(StringEqualWithContext(digest))
	return buf.String()
}

func VerifyComponent(resolver ocm.ComponentVersionResolver, signame, name string, digest string, other ...Option) {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	ExpectWithOffset(1, err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		VerifySignature(signame),
		Resolver(resolver),
		VerifyDigests(),
	)
	opts.Eval(other...)
	ExpectWithOffset(1, opts.Complete(cv.GetContext())).To(Succeed())
	dig, err := Apply(nil, nil, cv, opts)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, dig.Value).To(Equal(digest))
}

func FailVerifyComponent(resolver ocm.ComponentVersionResolver, signame, name string, digest string, msg string, other ...Option) {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	ExpectWithOffset(1, err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		VerifySignature(signame),
		Resolver(resolver),
		VerifyDigests(),
	)
	opts.Eval(other...)
	ExpectWithOffset(1, opts.Complete(cv.GetContext())).To(Succeed())
	_, err = Apply(nil, nil, cv, opts)
	ExpectWithOffset(1, err).To(MatchError(msg))
}

func Check(f func() error) {
	ExpectWithOffset(1, f()).To(Succeed())
}

func CheckResourceDigests(cd *compdesc.ComponentDescriptor, digests map[string]*metav1.DigestSpec, offsets ...int) {
	o := 4
	for _, a := range offsets {
		o += a
	}
	for i, r := range cd.Resources {
		By(fmt.Sprintf("resource %d", i), func() {
			if none.IsNone(r.Access.GetKind()) {
				ExpectWithOffset(o, r.Digest).To(BeNil())
			} else {
				ExpectWithOffset(o, r.Digest).NotTo(BeNil())
				if digests != nil {
					ExpectWithOffset(o, r.Digest).To(Equal(digests[r.Name]))
				}
			}
		})
	}
}

////////////////////////////////////////////////////////////////////////////////

const wrongDigest = "0a835d52867572bdaf7da7fb35ee59ad45c3db2dacdeeca62178edd5d07ef08c" // any wrong value

type EntryCheck interface {
	Mode() string
	Check1CheckD(cvd ocm.ComponentVersionAccess, d Substitutions)
	Check1CheckA(cva ocm.ComponentVersionAccess, d *metav1.DigestSpec, mopts ...ocm.ModificationOption)
	Check1Corrupt(cva ocm.ComponentVersionAccess, f *Finalizer, cvd ocm.ComponentVersionAccess)

	Check2Ref(cv ocm.ComponentVersionAccess, name string, d string)
}

type EntryLocal struct{}

func (*EntryLocal) Mode() string {
	return DIGESTMODE_LOCAL
}

func (*EntryLocal) Check1CheckD(cvd ocm.ComponentVersionAccess, _ Substitutions) {
	ExpectWithOffset(1, cvd.GetDescriptor().NestedDigests).To(BeNil())
}

func (*EntryLocal) Check1CheckA(cva ocm.ComponentVersionAccess, d *metav1.DigestSpec, _ ...ocm.ModificationOption) {
	CheckResourceDigests(cva.GetDescriptor(), map[string]*metav1.DigestSpec{
		"data_a": d,
	}, 1)
}

func (*EntryLocal) Check1Corrupt(cva ocm.ComponentVersionAccess, f *Finalizer, _ ocm.ComponentVersionAccess) {
	cva.GetDescriptor().Resources[0].Digest.Value = wrongDigest
	MustBeSuccessful(cva.Update())
	Check(f.Finalize)
}

func (*EntryLocal) Check2Ref(cv ocm.ComponentVersionAccess, name string, d string) {
	CheckCompRef(cv, name, CompDigestSpec(d), 1)
}

//////////

type EntryTop struct {
	found int
}

func (*EntryTop) Mode() string {
	return DIGESTMODE_TOP
}

func (*EntryTop) Check1CheckD(cvd ocm.ComponentVersionAccess, digests Substitutions) {
	ExpectWithOffset(1, cvd.GetDescriptor().NestedDigests).NotTo(BeNil())
	ExpectWithOffset(1, cvd.GetDescriptor().NestedDigests.String()).To(StringEqualTrimmedWithContext(`
github.com/mandelsoft/ref:v1: SHA-256:${D_COMPB}[jsonNormalisation/v1]
  data_b:v1[]: SHA-256:${D_DATAB}[genericBlobDigest/v1]
github.com/mandelsoft/ref2:v1: SHA-256:${D_COMPC}[jsonNormalisation/v1]
  data_c:v1[]: SHA-256:${D_DATAC}[genericBlobDigest/v1]
github.com/mandelsoft/test:v1: SHA-256:${D_COMPA}[jsonNormalisation/v1]
  data_a:v1[]: SHA-256:${D_DATAA}[genericBlobDigest/v1]
`, digests))
}

func (*EntryTop) Check1CheckA(cva ocm.ComponentVersionAccess, d *metav1.DigestSpec, mopts ...ocm.ModificationOption) {
	if ocm.NewModificationOptions(mopts...).IsSkipDigest() {
		ExpectWithOffset(1, cva.GetDescriptor().Resources[0].Digest).To(BeNil())
	} else {
		CheckResourceDigests(cva.GetDescriptor(), map[string]*metav1.DigestSpec{
			"data_a": d,
		}, 1)
	}
}

func (e *EntryTop) Check1Corrupt(_ ocm.ComponentVersionAccess, _ *Finalizer, cvd ocm.ComponentVersionAccess) {
	e.found = -1
	for i, n := range cvd.GetDescriptor().NestedDigests {
		if n.Name == COMPONENTA {
			n.Resources[0].Digest.Value = wrongDigest
			e.found = i
		}
	}
	ExpectWithOffset(1, e.found).NotTo(Equal(-1))
}

func (*EntryTop) Check2Ref(cv ocm.ComponentVersionAccess, name string, d string) {
	CheckCompRef(cv, name, nil, 1)
}
