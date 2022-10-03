// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package signing_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	. "github.com/open-component-model/ocm/pkg/env/builder"

	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
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
const OCINAMESPACE = "ocm/value"
const OCINAMESPACE2 = "ocm/ref"
const OCIVERSION = "v2.0"
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
			env.OCIContext().SetAlias(OCIHOST, ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCIPATH, accessio.PathFileSystem(env.FileSystem())))

			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				env.Namespace(OCINAMESPACE, func() {
					env.Manifest(OCIVERSION, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "manifestlayer")
						})
					})
				})
				env.Namespace(OCINAMESPACE2, func() {
					env.Manifest(OCIVERSION, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "otherlayer")
						})
					})
				})
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
								ociartefact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
							)
							env.Label("transportByValue", true)
						})
						env.Resource("ref", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								ociartefact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE2, OCIVERSION)),
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
			logrus.Infof("%+v\n", dig)
			Expect(dig.Value).To(Equal(digest))

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
			logrus.Infof("%+v\n", dig)
			Expect(dig.Value).To(Equal(digest))

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
			logrus.Infof("%+v\n", dig)
			Expect(dig.Value).To(Equal(digest))

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
			Expect(err.Error()).To(Equal("no signature found in github.com/mandelsoft/test:v1"))
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
			Expect(err.Error()).To(Equal("failed resolving component reference ref[github.com/mandelsoft/test:v1] in github.com/mandelsoft/ref:v1: component version \"github.com/mandelsoft/test:v1\" not found: oci artefact \"v1\" not found in component-descriptors/github.com/mandelsoft/test"))
		})
	})
})
