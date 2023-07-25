// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/common/accessobj"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/artdesc"
	. "github.com/open-component-model/ocm/v2/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/accessmethods/ociartifact"
	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	ctfocm "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/v2/pkg/mime"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var ldesc *artdesc.Descriptor

	_ = ldesc
	BeforeEach(func() {
		env = NewTestEnv()

		FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env.Builder)
			OCIManifest2(env.Builder)
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
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers ctf", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "ctf", ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring component "github.com/mandelsoft/test"...
  transferring version "github.com/mandelsoft/test:v1"...
  ...resource 0...
  ...adding component version...
`))

		Expect(env.DirExists(OUT)).To(BeTrue())

		tgt, err := ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, OUT, 0, accessio.PathFileSystem(env.FileSystem()))
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
		Expect(string(data)).To(Equal("{\"imageReference\":\"alias.alias/ocm/ref:v2.0\",\"type\":\"" + ociartifact.Type + "\"}"))

		data, err = json.Marshal(comp.GetDescriptor().Resources[1].Access)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"" + ociartifact.Type + "\"}"))
	})
})
