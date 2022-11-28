// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	tenv "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

var DefaultContext = ocm.New()

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

var _ = Describe("access method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(tenv.NewEnvironment())
		env.RSAKeyPair(SIGNATURE)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("valid", func() {
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

		It("sign flat version", func() {
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
			digest := "39ea26ac4391052a638319f64b8da2628acb51d304c3a1ac8f920a46f2d6dce7"
			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(closer.Close()).To(Succeed())
			Expect(archcloser.Close()).To(Succeed())
			Expect(dig.Value).To(StringEqualWithContext(digest))

			src, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err = src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digest))

			////////

			opts = NewOptions(
				VerifySignature(SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			dig, err = Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal(digest))

		})

		It("sign flat version with generic verification", func() {
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
			digest := "39ea26ac4391052a638319f64b8da2628acb51d304c3a1ac8f920a46f2d6dce7"
			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			closer.Close()
			archcloser.Close()
			Expect(dig.Value).To(StringEqualWithContext(digest))

			src, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err = src.LookupComponentVersion(COMPONENTA, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digest))

			////////

			opts = NewOptions(
				VerifySignature(),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			dig, err = Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal(digest))

		})

		It("sign deep version", func() {
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
				Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			digest := "05c4edd25661703e0c5caec8b0680c93738d8a8126d825adb755431fec29b7cb"
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())
			dig, err := Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			closer.Close()
			archcloser.Close()
			Expect(dig.Value).To(StringEqualWithContext(digest))

			src, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err = src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digest))

			////////

			opts = NewOptions(
				VerifySignature(SIGNATURE),
				Resolver(src),
				VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			dig, err = Apply(nil, nil, cv, opts)
			Expect(err).To(Succeed())
			Expect(dig.Value).To(Equal(digest))
		})

		It("fails generic verification", func() {
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
				VerifySignature(),
				Resolver(resolver),
				Update(), VerifyDigests(),
			)
			Expect(opts.Complete(signingattr.Get(DefaultContext))).To(Succeed())

			_, err = Apply(nil, nil, cv, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(StringEqualWithContext("github.com/mandelsoft/test:v1: no signature found"))
		})
	})

	Context("invalid", func() {
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

		It("fails signing version with unknow ref", func() {
			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)

			opts := NewOptions(
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
		})
	})
})
