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

package transfer_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCINAMESPACE = "ocm/value"
const OCINAMESPACE2 = "ocm/ref"
const OCIVERSION = "v2.0"
const OCIHOST = "alias"

func Check(env *TestEnv, ldesc *artdesc.Descriptor, out string) {
	tgt, err := ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, out, 0, accessio.PathFileSystem(env.FileSystem()))
	Expect(err).To(Succeed())
	defer tgt.Close()

	list, err := tgt.ComponentLister().GetComponents("", true)
	Expect(err).To(Succeed())
	Expect(list).To(Equal([]string{COMPONENT}))
	comp, err := tgt.LookupComponentVersion(COMPONENT, VERSION)
	Expect(err).To(Succeed())
	Expect(len(comp.GetDescriptor().Resources)).To(Equal(3))

	data, err := json.Marshal(comp.GetDescriptor().Resources[2].Access)
	Expect(err).To(Succeed())
	Expect(string(data)).To(Equal("{\"localReference\":\"sha256:f6a519fb1d0c8cef5e8d7811911fc7cb170462bbce19d6df067dae041250de7f\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/ref:v2.0\",\"type\":\"localBlob\"}"))

	data, err = json.Marshal(comp.GetDescriptor().Resources[1].Access)
	Expect(err).To(Succeed())
	Expect(string(data)).To(Equal("{\"localReference\":\"sha256:018520b2b249464a83e370619f544957b7936dd974468a128545eab88a0f53ed\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}"))

	racc, err := comp.GetResourceByIndex(1)
	Expect(err).To(Succeed())
	reader, err := ocm.ResourceReader(racc)
	Expect(err).To(Succeed())
	defer reader.Close()
	set, err := artefactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(reader))
	Expect(err).To(Succeed())
	defer set.Close()

	blob, err := set.GetBlob(ldesc.Digest)
	Expect(err).To(Succeed())
	data, err = blob.Get()
	Expect(err).To(Succeed())
	Expect(string(data)).To(Equal("manifestlayer"))
}

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var ldesc *artdesc.Descriptor

	_ = ldesc
	BeforeEach(func() {
		env = NewTestEnv()
		env.OCIContext().SetAlias(OCIHOST, ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCIPATH, accessio.PathFileSystem(env.FileSystem())))

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			env.Namespace(OCINAMESPACE, func() {
				env.Manifest(OCIVERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					ldesc = env.Layer(func() {
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
			env.Component(COMPONENT, func() {
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
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers ctf", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--resourcesByValue", ARCH, ARCH, OUT)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0...
...resource 1...
...resource 2...
...adding component version...
1 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())
		Check(env, ldesc, OUT)
	})

	It("transfers ctf to tgz", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--resourcesByValue", ARCH, ARCH, accessio.FormatTGZ.String()+"::"+OUT)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0...
...resource 1...
...resource 2...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		Check(env, ldesc, OUT)
	})

	It("transfers ctf to ctf+tgz", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--resourcesByValue", ARCH, ARCH, "ctf+"+accessio.FormatTGZ.String()+"::"+OUT)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0...
...resource 1...
...resource 2...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		Check(env, ldesc, OUT)
	})
})
