package transfer_test

import (
	"bytes"
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/oci/testhelper"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH      = "/tmp/ctf"
	PROVIDER  = "mandelsoft"
	VERSION   = "v1"
	COMPONENT = "github.com/mandelsoft/test"
	OUT       = "/tmp/res"
	OCIPATH   = "/tmp/oci"
	OCIHOST   = "alias"
)

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
  ...resource 0 testdata[PlainText]...
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
