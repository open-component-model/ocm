package download_test

import (
	"bytes"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/datacontext"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/resolvers"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	ARCH     = "/tmp/ca"
	VERSION  = "v1"
	COMP     = "test.de/x"
	COMP2    = "test.de/y"
	PROVIDER = "mandelsoft"
	OUT      = "/tmp/res"
)

const (
	PUBKEY  = "/tmp/pub"
	PRIVKEY = "/tmp/priv"
)

const (
	SIGNATURE = "test"
	SIGN_ALGO = rsa.Algorithm
)

const VERIFIED_FILE = "verified.yaml"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists single resource in ctf file", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "resources", "-O", OUT, ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: 8 byte(s) written
`))
		Expect(env.FileExists(OUT)).To(BeTrue())
		Expect(env.ReadFile(OUT)).To(Equal([]byte("testdata")))
	})

	It("registers download handler without config", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "resources", "--downloader", "helm/artifact:helm/v1", "-O", OUT, ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: 8 byte(s) written
`))
		Expect(env.FileExists(OUT)).To(BeTrue())
		Expect(env.ReadFile(OUT)).To(Equal([]byte("testdata")))
	})

	Context("with closure", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
				})
				env.Component(COMP2, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("moredata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "moredata")
						})
						env.Resource("otherdata", "", "PlainText", metav1.LocalRelation, func() {
							env.ExtraIdentity("id", "test")
							env.BlobStringData(mime.MIME_TEXT, "otherdata")
						})
						env.Reference("base", COMP, VERSION)
					})
				})
			})
		})

		It("downloads multiple files", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("download", "resources", "-O", OUT, "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
/tmp/res/test.de/y/v1/moredata: 8 byte(s) written
/tmp/res/test.de/y/v1/otherdata-id=test: 9 byte(s) written
`))

			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(Equal([]byte("moredata")))
			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(Equal([]byte("otherdata")))
		})

		It("downloads closure", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("download", "resources", "-r", "-O", OUT, "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
/tmp/res/test.de/y/v1/moredata: 8 byte(s) written
/tmp/res/test.de/y/v1/otherdata-id=test: 9 byte(s) written
/tmp/res/test.de/y/v1/test.de/x/v1/testdata: 8 byte(s) written
`))

			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(Equal([]byte("moredata")))
			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(Equal([]byte("otherdata")))
			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/"+COMP+"/"+VERSION+"/testdata"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/"+COMP+"/"+VERSION+"/testdata"))).To(Equal([]byte("testdata")))
		})
	})

	Context("verification", func() {
		priv, pub, err := rsa.Handler{}.CreateKeyPair()
		Expect(err).To(Succeed())

		BeforeEach(func() {
			data, err := rsa.KeyData(pub)
			Expect(err).To(Succeed())
			Expect(vfs.WriteFile(env.FileSystem(), PUBKEY, data, os.ModePerm)).To(Succeed())
			data, err = rsa.KeyData(priv)
			Expect(err).To(Succeed())
			Expect(vfs.WriteFile(env.FileSystem(), PRIVKEY, data, os.ModePerm)).To(Succeed())

			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
				})
			})
		})

		Prepare := func() {
			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
			defer Close(src, "repo")

			resolver := resolvers.NewCompoundResolver(src)
			cv := Must(resolver.LookupComponentVersion(COMP, VERSION))
			defer Close(cv, "cv")

			opts := signing.NewOptions(
				signing.Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
				signing.Resolver(resolver),
				signing.PrivateKey(SIGNATURE, priv),
				signing.Update(), signing.VerifyDigests(),
			)
			Expect(opts.Complete(env.OCMContext())).To(Succeed())

			Must(signing.Apply(nil, nil, cv, opts))
		}

		It("verifies download after component verification", func() {
			buf := bytes.NewBuffer(nil)

			Prepare()

			session := datacontext.NewSession()
			defer session.Close()

			Expect(env.CatchOutput(buf).Execute("verify", "components", "--verified", VERIFIED_FILE, "-V", "-s", SIGNATURE, "-k", PUBKEY, "--repo", ARCH, COMP+":"+VERSION)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "test.de/x:v1"[test.de/x:v1]...
  resource 0:  "name"="testdata": digest SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]
successfully verified test.de/x:v1 (digest SHA-256:ba5b4af72fcb707a4a8ebc48b088c0d8ed772cb021995b9b1fc7a01fbac29cd2)
`))

			buf.Reset()
			Expect(env.CatchOutput(buf).Execute("download", "resources", "--verified", VERIFIED_FILE, "-O", OUT, ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
/tmp/res: 8 byte(s) written
/tmp/res: resource content verified
`))
			Expect(env.FileExists(OUT)).To(BeTrue())
			Expect(env.ReadFile(OUT)).To(Equal([]byte("testdata")))
		})

		It("detects manipulated download after component verification", func() {
			buf := bytes.NewBuffer(nil)

			Prepare()

			session := datacontext.NewSession()
			defer session.Close()

			Expect(env.CatchOutput(buf).Execute("verify", "components", "--verified", VERIFIED_FILE, "-V", "-s", SIGNATURE, "-k", PUBKEY, "--repo", ARCH, COMP+":"+VERSION)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "test.de/x:v1"[test.de/x:v1]...
  resource 0:  "name"="testdata": digest SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]
successfully verified test.de/x:v1 (digest SHA-256:ba5b4af72fcb707a4a8ebc48b088c0d8ed772cb021995b9b1fc7a01fbac29cd2)
`))

			store := Must(signing.NewVerifiedStore(VERIFIED_FILE, env.FileSystem()))

			cd := store.Get(common.NewNameVersion(COMP, VERSION))

			cd.Resources[0].Digest.Value = "b" + cd.Resources[0].Digest.Value[1:]
			store.Add(cd)
			MustBeSuccessful(store.Save())

			ExpectError(env.Execute("download", "resources", "--verified", VERIFIED_FILE, "-O", OUT, ARCH)).To(MatchError("component version test.de/x:v1 corrupted"))
		})
	})
})
