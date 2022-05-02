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

package standard_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociregistry"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCINAMESPACE = "ocm/test"
const OCIVERSION = "v2.0"
const OCIHOST = "alias"

var _ = Describe("Transfer handler", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("it should copy a resource by value to a ctf file", func() {
		var ldesc *artdesc.Descriptor
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
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
					env.Resource("artefact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociregistry.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})

		env.OCIContext().SetAlias(OCIHOST, ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCIPATH, accessio.PathFileSystem(env.FileSystem())))

		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()
		handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, src, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		list, err := tgt.ComponentLister().GetComponents("", true)
		Expect(err).To(Succeed())
		Expect(list).To(Equal([]string{COMPONENT}))
		comp, err := tgt.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
		data, err := json.Marshal(comp.GetDescriptor().Resources[1].Access)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"localReference\":\"sha256:018520b2b249464a83e370619f544957b7936dd974468a128545eab88a0f53ed\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"type\":\"localBlob\"}"))

		r, err := comp.GetResourceByIndex(1)
		Expect(err).To(Succeed())
		meth, err := r.AccessMethod()
		Expect(err).To(Succeed())
		defer meth.Close()
		reader, err := meth.Reader()
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
	})
})
