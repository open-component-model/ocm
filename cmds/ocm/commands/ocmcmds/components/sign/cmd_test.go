// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package sign_test

import (
	"bytes"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

const ARCH = "/tmp/ctf"
const ARCH2 = "/tmp/ctf2"
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

const D_COMPONENTA = "01de99400030e8336020059a435cea4e7fe8f21aad4faf619da882134b85569d"
const D_COMPONENTB = "5f416ec59629d6af91287e2ba13c6360339b6a0acf624af2abd2a810ce4aefce"

var substitutions = Substitutions{
	"test": D_COMPONENTA,
	"r0":   D_TESTDATA,
	"r1":   DS_OCIMANIFEST1.Value,
	"r2":   DS_OCIMANIFEST2.Value,
	"ref":  D_COMPONENTB,
	"rb0":  D_OTHERDATA,
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

		It("sign single component in component archive", func() {
			prepareEnv(env, ARCH, "")

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTA+":"+VERSION)).To(Succeed())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
applying to version "github.com/mandelsoft/test:v1"[github.com/mandelsoft/test:v1]...
  resource 0:  "name"="testdata": digest SHA-256:${r0}[genericBlobDigest/v1]
  resource 1:  "name"="value": digest SHA-256:${r1}[ociArtifactDigest/v1]
  resource 2:  "name"="ref": digest SHA-256:${r2}[ociArtifactDigest/v1]
successfully signed github.com/mandelsoft/test:v1 (digest SHA-256:${test})`,
				substitutions),
			)

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPONENTA))
		})

		It("sign component archive", func() {
			prepareEnv(env, ARCH, ARCH)

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(Succeed())

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
`, substitutions))

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPONENTB))
		})

		It("sign component archive with --lookup option", func() {
			prepareEnv(env, ARCH2, ARCH)

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("sign", "components", "--lookup", ARCH2, "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(Succeed())

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
`, substitutions))

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(D_COMPONENTB))
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
Error: signing: github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: component "github.com/mandelsoft/test" not found in ComponentArchive
`))
		})

		It("sign archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, ARCH)).To(HaveOccurred())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Error: signing: github.com/mandelsoft/ref:v1: failed resolving component reference ref[github.com/mandelsoft/test:v1]: component "github.com/mandelsoft/test" not found in ComponentArchive
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
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=acme.org", "--cakey", "root.priv", "--cacert", "root.cert", "ca.priv")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair ca.priv[ca.cert]
`))
		Expect(env.FileExists("ca.priv")).To(BeTrue())
		Expect(env.FileExists("ca.cert")).To(BeTrue())

		// create signing vcertificate from CA
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=mandelsoft", "C=DE", "--cakey", "ca.priv", "--cacert", "ca.cert", "--rootcerts", "root.cert", "key.priv")).To(Succeed())
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
		Expect(env.CatchOutput(buf).Execute("sign", "component", ARCH, "-K", "key.priv", "-k", "key.cert", "--ca-cert", "root.cert", "-s", "mandelsoft", "-I", "CN=mandelsoft")).To(Succeed())
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
})

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
