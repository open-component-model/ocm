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

package sign_test

import (
	"bytes"
	"os"

	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

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

const PUBKEY = "/tmp/pub"
const PRIVKEY = "/tmp/priv"

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

		It("sign component archive", func() {
			buf := bytes.NewBuffer(nil)
			digest := "05c4edd25661703e0c5caec8b0680c93738d8a8126d825adb755431fec29b7cb"
			Expect(env.CatchOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, "--repo", ARCH, COMPONENTB+":"+VERSION)).To(Succeed())

			Expect("\n" + buf.String()).To(Equal(`
applying to version "github.com/mandelsoft/ref:v1"...
  applying to version "github.com/mandelsoft/test:v1"...
    resource 0:  "name"="testdata": digest sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50[genericBlobDigest/v1]
    resource 1:  "name"="value": digest sha256:0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f[ociArtifactDigest/v1]
    resource 2:  "name"="ref": digest sha256:c2d2dca275c33c1270dea6168a002d67c0e98780d7a54960758139ae19984bd7[ociArtifactDigest/v1]
  reference 0:  github.com/mandelsoft/test:v1: digest sha256:39ea26ac4391052a638319f64b8da2628acb51d304c3a1ac8f920a46f2d6dce7[jsonNormalisation/v1]
  resource 0:  "name"="otherdata": digest sha256:54b8007913ec5a907ca69001d59518acfd106f7b02f892eabf9cae3f8b2414b4[genericBlobDigest/v1]
successfully signed github.com/mandelsoft/ref:v1 (digest sha256:` + digest + `)
`))

			session := datacontext.NewSession()
			defer session.Close()

			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			session.AddCloser(src)
			cv, err := src.LookupComponentVersion(COMPONENTB, VERSION)
			Expect(err).To(Succeed())
			session.AddCloser(cv)
			Expect(cv.GetDescriptor().Signatures[0].Digest.Value).To(Equal(digest))

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
			Expect("\n" + buf.String()).To(Equal(`
Error: {signing: failed resolving component reference "ref" [github.com/mandelsoft/test:v1] in github.com/mandelsoft/ref:v1: ocm reference "github.com/mandelsoft/test:v1" not found}
`))
		})

		It("sign archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, ARCH)).To(HaveOccurred())
			Expect("\n" + buf.String()).To(Equal(`
Error: {signing: failed resolving component reference "ref" [github.com/mandelsoft/test:v1] in github.com/mandelsoft/ref:v1: ocm reference "github.com/mandelsoft/test:v1" not found}
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
			Expect("\n" + buf.String()).To(Equal(`
Error: {signing: failed resolving component reference "ref" [github.com/mandelsoft/test:v1] in github.com/mandelsoft/ref:v1: ocm reference "github.com/mandelsoft/test:v1" not found}
`))
		})

		It("sign archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchErrorOutput(buf).Execute("sign", "components", "-s", SIGNATURE, "-K", PRIVKEY, ARCH)).To(HaveOccurred())
			Expect("\n" + buf.String()).To(Equal(`
Error: {signing: failed resolving component reference "ref" [github.com/mandelsoft/test:v1] in github.com/mandelsoft/ref:v1: ocm reference "github.com/mandelsoft/test:v1" not found}
`))
		})
	})
})
