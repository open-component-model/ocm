// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/finalizer"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/none"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	tenv "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha512"
)

var DefaultContext = ocm.New()

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENTA = "github.com/mandelsoft/test"
const COMPONENTB = "github.com/mandelsoft/ref"
const COMPONENTC = "github.com/mandelsoft/ref2"
const COMPONENTD = "github.com/mandelsoft/top"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

const SIGNATURE = "test"
const SIGNATURE2 = "second"
const SIGN_ALGO = rsa.Algorithm

var _ = Describe("access method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(tenv.NewEnvironment(tenv.ModifiableTestData()))
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
			resolver := ocm.NewCompoundResolver(src)

			cv := Must(resolver.LookupComponentVersion(COMPONENTA, VERSION))
			closer := session.AddCloser(cv)

			digest := "123d48879559d16965a54eba9a3e845709770f4f0be984ec8db2f507aa78f338"

			pr, buf := common.NewBufferedPrinter()
			// key taken from signing attr
			dig := Must(SignComponentVersion(pr, cv, SIGNATURE, nil, SignerByName(SIGN_ALGO), Resolver(resolver), DigestMode(mode)))
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

			dig = Must(VerifyComponentVersion(pr, cv, SIGNATURE, nil, Resolver(resolver)))
			Expect(dig.Value).To(Equal(digest))
		},
			Entry(DIGESTMODE_TOP, DIGESTMODE_TOP),
			Entry(DIGESTMODE_LOCAL, DIGESTMODE_LOCAL),
		)
	})

	Context("valid", func() {
		digestA := "01de99400030e8336020059a435cea4e7fe8f21aad4faf619da882134b85569d"
		digestB := "5f416ec59629d6af91287e2ba13c6360339b6a0acf624af2abd2a810ce4aefce"

		localDigests := common.Properties{
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
			resolver := ocm.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			closer := session.AddCloser(cv)

			opts := NewOptions(
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			pr, buf := common.NewBufferedPrinter()
			dig, err := Apply(pr, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(closer.Close()).To(Succeed())
			Expect(archcloser.Close()).To(Succeed())
			Expect(dig.Value).To(StringEqualWithContext(digestA))

			src, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
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
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

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
			Expect(err.Error()).To(Equal("github.com/mandelsoft/test:v1: calculated resource digest ([{HashAlgorithm:SHA-256 NormalisationAlgorithm:genericBlobDigest/v1 Value:" + D_TESTDATA + "}]) mismatches existing digest (SHA-256:010ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]) for testdata:v1 (Local blob sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[])"))
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
			resolver := ocm.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			closer := session.AddCloser(cv)

			opts := NewOptions(
				DigestMode(mode),
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())
			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			closer.Close()
			archcloser.Close()
			Expect(dig.Value).To(StringEqualWithContext(digestA))

			src, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
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
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

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
			resolver := ocm.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			closer := session.AddCloser(cv)

			opts := NewOptions(
				DigestMode(mode),
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			pr, buf := common.NewBufferedPrinter()
			dig, err := Apply(pr, nil, cv, opts)
			Expect(err).To(Succeed())
			closer.Close()
			archcloser.Close()
			Expect(dig.Value).To(StringEqualWithContext(digestB))

			src, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
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
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

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
			resolver := ocm.NewCompoundResolver(src)

			cv, err := resolver.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)

			opts := NewOptions(
				DigestMode(mode),
				VerifySignature(),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

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
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(src),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

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

	Context("rhombus", func() {
		D_DATAA := "8a835d52867572bdaf7da7fb35ee59ad45c3db2dacdeeca62178edd5d07ef08c"
		D_DATAB := "5f103fcedc97b81bfc1841447d164781ed0f6244ce20b26d7a8a7d5880156c33"
		D_DATAB512 := "a9469fc2e9787c8496cf1526508ae86d4e855715ef6b8f7031bdc55759683762f1c330b94a4516dff23e32f19fb170cbcb53015f1ffc0d77624ee5c9a288a030"
		D_DATAC := "90e06e32c46338db42d78d49fee035063d4b10e83cfbf0d1831e14527245da12"
		D_DATAD := "5a5c3f681c2af10d682926a635a1dc9dfe7087d4fa3daf329bf0acad540911a9"

		DS_DATAA := TextResourceDigestSpec(D_DATAA)
		DS_DATAB := TextResourceDigestSpec(D_DATAB)
		DS_DATAC := TextResourceDigestSpec(D_DATAC)
		DS_DATAD := TextResourceDigestSpec(D_DATAD)

		D_COMPA := "bdb62ce8299f10e230b91bc9a9bfdbf2d33147f205fcf736d802c7e1cec7b5e8"
		D_COMPB := "d1def1b60cc8b241451b0e3fccb705a9d99db188b72ec4548519017921700857"
		D_COMPBR := "e47deeca35bc34116770a50a88954a0b028eb4e236d089b84e419c6d7ce15d97"
		D_COMPC := "b376a7b440c0b1e506e54a790966119a8e229cf9226980b84c628d77ef06fc58"
		D_COMPD := "64674d3e2843d36c603f44477e4cd66ee85fe1a91227bbcd271202429024ed61"
		D_COMPB512 := "08366761127c791e550d2082e34e68c8836739c68f018f969a46a17a6c13b529390303335ee0ae3cd938af9e0f31665427a1b45360622d864a5dbe053917a75d"

		localDigests := common.Properties{
			"D_DATAA":    D_DATAA,
			"D_DATAB":    D_DATAB,
			"D_DATAB512": D_DATAB512,
			"D_DATAC":    D_DATAC,
			"D_DATAD":    D_DATAD,

			"D_COMPA":    D_COMPA,
			"D_COMPB":    D_COMPB,
			"D_COMPBR":   D_COMPBR,
			"D_COMPC":    D_COMPC,
			"D_COMPD":    D_COMPD,
			"D_COMPB512": D_COMPB512,
		}

		_, _, _, _ = DS_DATAA, DS_DATAB, DS_DATAC, DS_DATAD

		BeforeEach(func() {
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
		})

		DescribeTable("hashes unsigned", func(c EntryCheck) {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			{
				arch := finalizer.Nested()
				src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
				Expect(err).To(Succeed())
				arch.Close(src)

				resolver := ocm.NewCompoundResolver(src)

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

			c.Check1CheckA(cva, DS_DATAA)

			////////

			VerifyHashes(src, COMPONENTD, D_COMPD)

			c.Check1Corrupt(cva, sub, cv)

			opts := NewOptions(
				Resolver(src),
				VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())
			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("github.com/mandelsoft/top:v1: failed applying to component reference refb[github.com/mandelsoft/ref:v1]: github.com/mandelsoft/top:v1->github.com/mandelsoft/ref:v1: failed applying to component reference ref[github.com/mandelsoft/test:v1]: github.com/mandelsoft/top:v1->github.com/mandelsoft/ref:v1->github.com/mandelsoft/test:v1: calculated resource digest ([{HashAlgorithm:SHA-256 NormalisationAlgorithm:genericBlobDigest/v1 Value:" + D_DATAA + "}]) mismatches existing digest (SHA-256:" + wrongDigest + "[genericBlobDigest/v1]) for data_a:v1 (Local blob sha256:" + D_DATAA + "[])"))
		},
			Entry(DIGESTMODE_TOP, &EntryTop{}),
			Entry(DIGESTMODE_LOCAL, &EntryLocal{}),
		)

		DescribeTable("signs unsigned", func(c EntryCheck) {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			{
				arch := finalizer.Nested()
				src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
				Expect(err).To(Succeed())
				arch.Close(src)

				resolver := ocm.NewCompoundResolver(src)

				log := SignComponent(resolver, COMPONENTD, D_COMPD, DigestMode(c.Mode()))

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

			c.Check1CheckA(cva, DS_DATAA)
			////////

			cvb := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
			finalizer.Close(cvb)
			Expect(len(cvb.GetDescriptor().Signatures)).To(Equal(0))
			c.Check2Ref(cvb, "ref", D_COMPA)

			cvc := Must(src.LookupComponentVersion(COMPONENTB, VERSION))
			finalizer.Close(cvc)
			Expect(len(cvb.GetDescriptor().Signatures)).To(Equal(0))
			c.Check2Ref(cvb, "ref", D_COMPA)

			VerifyComponent(src, COMPONENTD, D_COMPD)
		},
			Entry(DIGESTMODE_TOP, &EntryTop{}),
			Entry(DIGESTMODE_LOCAL, &EntryLocal{}),
		)

		It("signs and rehashes presigned in top mode", func() {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			// digestD := "f428e9af521fcbd3229b01fb5cc0c4875ddb199de6356acf40642df6917d0e8f"
			digestD := "342d30317bee13ec30d815122f23b19d9ee54a15ff8be1ec550c8072d5a6dba6"
			digestB := D_COMPB512
			{
				arch := finalizer.Nested()
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				arch.Close(src)

				resolver := ocm.NewCompoundResolver(src)

				SignComponent(resolver, COMPONENTB, digestB, DigestMode(DIGESTMODE_TOP), HashByAlgo(sha512.Algorithm))
				VerifyComponent(resolver, COMPONENTB, digestB)
				log := SignComponent(resolver, COMPONENTD, digestD, DigestMode(DIGESTMODE_TOP))

				Expect(log).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/top:v1"[github.com/mandelsoft/top:v1]...
  no digest found for "github.com/mandelsoft/ref:v1"
  applying to version "github.com/mandelsoft/ref:v1"[github.com/mandelsoft/top:v1]...
    no digest found for "github.com/mandelsoft/test:v1"
    applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/top:v1]...
      resource 0:  "name"="data_a": digest SHA-256:8a835d52867572bdaf7da7fb35ee59ad45c3db2dacdeeca62178edd5d07ef08c[genericBlobDigest/v1]
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:bdb62ce8299f10e230b91bc9a9bfdbf2d33147f205fcf736d802c7e1cec7b5e8[jsonNormalisation/v1]
    resource 0:  "name"="data_b": digest SHA-512:a9469fc2e9787c8496cf1526508ae86d4e855715ef6b8f7031bdc55759683762f1c330b94a4516dff23e32f19fb170cbcb53015f1ffc0d77624ee5c9a288a030[genericBlobDigest/v1]
  reference 0:  github.com/mandelsoft/ref:v1: digest SHA-256:e47deeca35bc34116770a50a88954a0b028eb4e236d089b84e419c6d7ce15d97[jsonNormalisation/v1]
  no digest found for "github.com/mandelsoft/ref2:v1"
  applying to version "github.com/mandelsoft/ref2:v1"[github.com/mandelsoft/top:v1]...
    reference 0:  github.com/mandelsoft/test:v1: digest SHA-256:bdb62ce8299f10e230b91bc9a9bfdbf2d33147f205fcf736d802c7e1cec7b5e8[jsonNormalisation/v1]
    resource 0:  "name"="data_c": digest SHA-256:90e06e32c46338db42d78d49fee035063d4b10e83cfbf0d1831e14527245da12[genericBlobDigest/v1]
  reference 1:  github.com/mandelsoft/ref2:v1: digest SHA-256:b376a7b440c0b1e506e54a790966119a8e229cf9226980b84c628d77ef06fc58[jsonNormalisation/v1]
  resource 0:  "name"="data_d": digest SHA-256:5a5c3f681c2af10d682926a635a1dc9dfe7087d4fa3daf329bf0acad540911a9[genericBlobDigest/v1]
`))
				Defer(arch.Finalize)
			}

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			finalizer.Close(src)
			cv := Must(src.LookupComponentVersion(COMPONENTD, VERSION))
			finalizer.Close(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digestD))

			Expect(cv.GetDescriptor().NestedDigests).NotTo(BeNil())
			Expect(cv.GetDescriptor().NestedDigests.String()).To(StringEqualTrimmedWithContext(`
github.com/mandelsoft/ref:v1: SHA-256:${D_COMPBR}[jsonNormalisation/v1]
  data_b:v1[]: SHA-512:${D_DATAB512}[genericBlobDigest/v1]
github.com/mandelsoft/ref2:v1: SHA-256:${D_COMPC}[jsonNormalisation/v1]
  data_c:v1[]: SHA-256:${D_DATAC}[genericBlobDigest/v1]
github.com/mandelsoft/test:v1: SHA-256:${D_COMPA}[jsonNormalisation/v1]
  data_a:v1[]: SHA-256:${D_DATAA}[genericBlobDigest/v1]
`, localDigests))

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			////////

			VerifyComponent(src, COMPONENTD, digestD)
		})

		It("verifies after sub level signing", func() {
			var finalizer Finalizer
			defer Defer(finalizer.Finalize)

			digestD := D_COMPD
			digestB := D_COMPB512
			{
				arch := finalizer.Nested()
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				arch.Close(src)

				resolver := ocm.NewCompoundResolver(src)

				fmt.Printf("SIGN D\n")
				log := SignComponent(resolver, COMPONENTD, digestD, DigestMode(DIGESTMODE_TOP))

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

				fmt.Printf("SIGN B\n")
				SignComponent(resolver, COMPONENTB, digestB, HashByAlgo(sha512.Algorithm), DigestMode(DIGESTMODE_TOP))
				fmt.Printf("VERIFY B\n")
				VerifyComponent(resolver, COMPONENTB, digestB)

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
`, localDigests))

			cva, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			finalizer.Close(cva)
			Expect(len(cva.GetDescriptor().Signatures)).To(Equal(0))

			////////
			fmt.Printf("VERIFY D\n")
			VerifyComponent(src, COMPONENTD, digestD)
		})

		It("fixes digest mode", func() {
			var finalizer Finalizer
			defer Check(finalizer.Finalize)

			{ // sign with mode local
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				finalizer.Close(src)

				resolver := ocm.NewCompoundResolver(src)

				fmt.Printf("SIGN B\n")
				_ = SignComponent(resolver, COMPONENTB, D_COMPB, DigestMode(DIGESTMODE_LOCAL))
				VerifyComponent(src, COMPONENTB, D_COMPB)
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

				resolver := ocm.NewCompoundResolver(src)

				fmt.Printf("RESIGN B\n")
				_ = SignComponent(resolver, COMPONENTB, D_COMPB, DigestMode(DIGESTMODE_TOP), SignatureName(SIGNATURE2, true))
				VerifyComponent(src, COMPONENTB, D_COMPB, SignatureName(SIGNATURE2, true))
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
		})

		It("fixes digest mode in recursive signing", func() {
			var finalizer Finalizer
			defer Check(finalizer.Finalize)

			{ // sign with mode local
				src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
				finalizer.Close(src)

				resolver := ocm.NewCompoundResolver(src)

				fmt.Printf("SIGN B\n")
				_ = SignComponent(resolver, COMPONENTB, D_COMPB, DigestMode(DIGESTMODE_TOP), SignatureName(SIGNATURE2, true))
				VerifyComponent(src, COMPONENTB, D_COMPB, SignatureName(SIGNATURE2, true))
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

				resolver := ocm.NewCompoundResolver(src)

				fmt.Printf("SIGN D\n")
				_ = SignComponent(resolver, COMPONENTD, D_COMPD, Recursive(), DigestMode(DIGESTMODE_LOCAL))
				VerifyComponent(src, COMPONENTD, D_COMPD)
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
		})
	})
})

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

func SignComponent(resolver ocm.ComponentVersionResolver, name string, digest string, other ...Option) string {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	Expect(err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
		Resolver(resolver),
		VerifyDigests(),
	)
	opts.Eval(other...)
	ExpectWithOffset(1, opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

	pr, buf := common.NewBufferedPrinter()
	dig, err := Apply(pr, nil, cv, opts)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, dig.Value).To(StringEqualWithContext(digest))
	return buf.String()
}

func VerifyComponent(resolver ocm.ComponentVersionResolver, name string, digest string, other ...Option) {
	cv, err := resolver.LookupComponentVersion(name, VERSION)
	ExpectWithOffset(1, err).To(Succeed())
	defer cv.Close()

	opts := NewOptions(
		VerifySignature(SIGNATURE),
		Resolver(resolver),
		VerifyDigests(),
	)
	opts.Eval(other...)
	ExpectWithOffset(1, opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())
	dig, err := Apply(nil, nil, cv, opts)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, dig.Value).To(Equal(digest))
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
	Check1CheckD(cvd ocm.ComponentVersionAccess, d common.Properties)
	Check1CheckA(cva ocm.ComponentVersionAccess, d *metav1.DigestSpec)
	Check1Corrupt(cva ocm.ComponentVersionAccess, f *Finalizer, cvd ocm.ComponentVersionAccess)

	Check2Ref(cv ocm.ComponentVersionAccess, name string, d string)
}

type EntryLocal struct{}

func (*EntryLocal) Mode() string {
	return DIGESTMODE_LOCAL
}

func (*EntryLocal) Check1CheckD(cvd ocm.ComponentVersionAccess, _ common.Properties) {
	ExpectWithOffset(1, cvd.GetDescriptor().NestedDigests).To(BeNil())
}

func (*EntryLocal) Check1CheckA(cva ocm.ComponentVersionAccess, d *metav1.DigestSpec) {
	CheckResourceDigests(cva.GetDescriptor(), map[string]*metav1.DigestSpec{
		"data_a": d,
	}, 1)
}

func (*EntryLocal) Check1Corrupt(cva ocm.ComponentVersionAccess, f *Finalizer, _ ocm.ComponentVersionAccess) {
	cva.GetDescriptor().Resources[0].Digest.Value = wrongDigest
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

func (*EntryTop) Check1CheckD(cvd ocm.ComponentVersionAccess, digests common.Properties) {
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

func (*EntryTop) Check1CheckA(cva ocm.ComponentVersionAccess, _ *metav1.DigestSpec) {
	ExpectWithOffset(1, cva.GetDescriptor().Resources[0].Digest).To(BeNil())
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
