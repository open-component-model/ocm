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

package spiff_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ctf"
const ARCH2 = "/tmp/ctf2"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const COMPONENT2 = "github.com/mandelsoft/test2"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

const script1 = `
rules:
  resource:
    <<: (( &template ))
    process: (( values.component.name == "github.com/mandelsoft/test" ))
  componentversion:
    <<: (( &template ))
    process: 
      process: true
      script: (( rules.default ))
      repospec:
        type: dummy
  default:
    <<: (( &template ))
    process: false

process: (( (*(rules[mode] || rules.default)).process ))
`

var _ = Describe("Transfer handler", func() {

	It("handles bool", func() {

		handler, err := spiff.New(spiff.Script([]byte(script1)))
		Expect(err).To(Succeed())

		binding := map[string]interface{}{
			"component": map[string]interface{}{
				"name":    COMPONENT,
				"version": VERSION,
			},
		}
		ok, err := handler.(*spiff.Handler).EvalBool("resource", binding, "process")
		Expect(err).To(Succeed())
		Expect(ok).To(BeTrue())
	})

	It("handles componentversion", func() {

		handler, err := spiff.New(spiff.Script([]byte(script1)))
		Expect(err).To(Succeed())

		binding := map[string]interface{}{
			"component": map[string]interface{}{
				"name":    COMPONENT,
				"version": VERSION,
			},
		}
		ok, r, s, err := handler.(*spiff.Handler).EvalRecursion("componentversion", binding, "process")
		Expect(err).To(Succeed())
		Expect(ok).To(BeTrue())
		Expect(string(s)).To(Equal("process: false\n"))
		Expect(string(r)).To(Equal("{\"type\":\"dummy\"}"))
	})

	It("handles simple componentversion", func() {

		handler, err := spiff.New(spiff.Script([]byte(script1)))
		Expect(err).To(Succeed())
		binding := map[string]interface{}{
			"component": map[string]interface{}{
				"name":    COMPONENT,
				"version": VERSION,
			},
		}
		ok, r, s, err := handler.(*spiff.Handler).EvalRecursion("resource", binding, "process")
		Expect(err).To(Succeed())
		Expect(ok).To(BeTrue())
		Expect(r).To(BeNil())
		Expect(s).To(BeNil())
	})

	Context("handler", func() {
		var env *Builder
		var ldesc *artdesc.Descriptor

		BeforeEach(func() {
			env = NewBuilder(NewEnvironment())

			FakeOCIRepo(env, OCIPATH, OCIHOST)

			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				ldesc = OCIManifest1(env)
				OCIManifest2(env)
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

			env.OCMCommonTransport(ARCH2, accessio.FormatDirectory, func() {
				env.Component(COMPONENT2, func() {
					env.Version(VERSION, func() {
						env.Reference("ref", COMPONENT, VERSION)
						env.Provider(PROVIDER)
					})
				})
			})
		})

		AfterEach(func() {
			env.Cleanup()
		})

		const script2 = `
rules:
  resource:
    <<: (( &template ))
#    process: (( values.element.name == "value" ))
    process: (( values.element.labels.transportByValue.value || false ))

  default:
    <<: (( &template ))
    process: false

process: (( (*(rules[mode] || rules.default)).process ))
`

		It("it should copy all resource by value to a ctf file without script", func() {
			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
			tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
			Expect(err).To(Succeed())
			defer tgt.Close()

			handler, err := spiff.New(standard.ResourcesByValue())
			Expect(err).To(Succeed())
			err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
			Expect(err).To(Succeed())
			Expect(env.DirExists(OUT)).To(BeTrue())

			list, err := tgt.ComponentLister().GetComponents("", true)
			Expect(err).To(Succeed())
			Expect(list).To(Equal([]string{COMPONENT}))
			comp, err := tgt.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(3))

			data, err := json.Marshal(comp.GetDescriptor().Resources[2].Access)
			Expect(err).To(Succeed())
			fmt.Printf("%s\n", string(data))
			hash := HashManifest2(artefactset.IsOCIDefaultFormat())
			Expect(string(data)).To(Equal("{\"localReference\":\"" + hash + "\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"" + OCINAMESPACE2 + ":" + OCIVERSION + "\",\"type\":\"localBlob\"}"))

			data, err = json.Marshal(comp.GetDescriptor().Resources[1].Access)
			Expect(err).To(Succeed())
			fmt.Printf("%s\n", string(data))
			hash = HashManifest1(artefactset.IsOCIDefaultFormat())
			Expect(string(data)).To(Equal("{\"localReference\":\"" + hash + "\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"" + OCINAMESPACE + ":" + OCIVERSION + "\",\"type\":\"localBlob\"}"))

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
		})

		It("it should copy one resource by value to a ctf file with script", func() {
			src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
			tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
			Expect(err).To(Succeed())
			defer tgt.Close()

			handler, err := spiff.New(standard.ResourcesByValue(), spiff.Script([]byte(script2)))
			Expect(err).To(Succeed())
			err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
			Expect(err).To(Succeed())
			Expect(env.DirExists(OUT)).To(BeTrue())

			list, err := tgt.ComponentLister().GetComponents("", true)
			Expect(err).To(Succeed())
			Expect(list).To(Equal([]string{COMPONENT}))
			comp, err := tgt.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(3))

			// index 2: by ref
			data, err := json.Marshal(comp.GetDescriptor().Resources[2].Access)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("{\"imageReference\":\"" + oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE2, OCIVERSION) + "\",\"type\":\"" + ociartefact.Type + "\"}"))

			// index 1: by value
			data, err = json.Marshal(comp.GetDescriptor().Resources[1].Access)
			Expect(err).To(Succeed())
			hash := HashManifest1(artefactset.IsOCIDefaultFormat())
			Expect(string(data)).To(Equal("{\"localReference\":\"" + hash + "\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"" + OCINAMESPACE + ":" + OCIVERSION + "\",\"type\":\"localBlob\"}"))

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
		})

		It("it should use additional resolver to resolve component ref", func() {
			parentSrc, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH2, 0, env)
			Expect(err).To(Succeed())
			cv, err := parentSrc.LookupComponentVersion(COMPONENT2, VERSION)
			Expect(err).To(Succeed())
			childSrc, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
			Expect(err).To(Succeed())
			tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
			Expect(err).To(Succeed())
			defer tgt.Close()
			handler, err := standard.New(standard.Recursive(), standard.Resolver(childSrc))
			Expect(err).To(Succeed())
			err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
			Expect(err).To(Succeed())
			Expect(env.DirExists(OUT)).To(BeTrue())

			list, err := tgt.ComponentLister().GetComponents("", true)
			Expect(err).To(Succeed())
			Expect(list).To(ContainElements([]string{COMPONENT2, COMPONENT}))
			_, err = tgt.LookupComponentVersion(COMPONENT2, VERSION)
			Expect(err).To(Succeed())
			_, err = tgt.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
		})
	})
})
