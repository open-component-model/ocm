package transfer_test

import (
	"bytes"
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	ocmutils "ocm.software/ocm/api/ocm/ocmutils"
	handlercfg "ocm.software/ocm/api/ocm/tools/transfer/transferhandler/config"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH       = "/tmp/ctf"
	ARCH2      = "/tmp/ctf2"
	PROVIDER   = "mandelsoft"
	VERSION    = "v1"
	COMPONENT  = "github.com/mandelsoft/test"
	COMPONENT2 = "github.com/mandelsoft/test2"
	OUT        = "/tmp/res"
	OCIPATH    = "/tmp/oci"
	OCIHOST    = "alias"
)

func CheckComponentInArchive(env *TestEnv, ldesc *artdesc.Descriptor, out string) {
	tgt, err := ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, out, 0, accessio.PathFileSystem(env.FileSystem()))
	Expect(err).To(Succeed())
	defer Close(tgt, "ctf")

	list, err := tgt.ComponentLister().GetComponents("", true)
	Expect(err).To(Succeed())
	Expect(list).To(Equal([]string{COMPONENT}))
	CheckComponent(env, ldesc, tgt)
}

func CheckComponent(env *TestEnv, ldesc *artdesc.Descriptor, tgt ocm.Repository) {
	comp, err := tgt.LookupComponentVersion(COMPONENT, VERSION)
	Expect(err).To(Succeed())
	defer Close(comp, "comvers")
	Expect(len(comp.GetDescriptor().Resources)).To(Equal(3))

	data, err := json.Marshal(comp.GetDescriptor().Resources[2].Access)
	Expect(err).To(Succeed())
	hash := HashManifest2(artifactset.DefaultArtifactSetDescriptorFileName)
	Expect(string(data)).To(StringEqualWithContext("{\"localReference\":\"" + hash + "\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/ref:v2.0\",\"type\":\"localBlob\"}"))

	data, err = json.Marshal(comp.GetDescriptor().Resources[1].Access)
	Expect(err).To(Succeed())
	hash = HashManifest1(artifactset.DefaultArtifactSetDescriptorFileName)
	Expect(string(data)).To(StringEqualWithContext("{\"localReference\":\"" + hash + "\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}"))

	racc, err := comp.GetResourceByIndex(1)
	Expect(err).To(Succeed())
	reader, err := ocmutils.GetResourceReader(racc)
	Expect(err).To(Succeed())
	defer reader.Close()
	set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(reader))
	Expect(err).To(Succeed())
	defer set.Close()

	_, blob, err := set.GetBlobData(ldesc.Digest)
	Expect(err).To(Succeed())
	defer Close(blob, "blob")
	data, err = blob.Get()
	Expect(err).To(Succeed())
	Expect(string(data)).To(Equal("manifestlayer"))
}

var _ = Describe("Test Environment", func() {
	var (
		env   *TestEnv
		ldesc *artdesc.Descriptor
	)

	_ = ldesc
	BeforeEach(func() {
		env = NewTestEnv()

		FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			ldesc = OCIManifest1(env.Builder)
			OCIManifest2(env.Builder)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
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

		env.OCMCommonTransport(ARCH2, accessio.FormatDirectory, func() {
			env.Component(COMPONENT2, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("otherdate", "", resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "otherdata")
					})
					env.Reference("ref", COMPONENT, VERSION)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers ctf", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", ARCH, ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[plainText]...
...resource 1 value[ociImage](ocm/value:v2.0)...
...resource 2 ref[ociImage](ocm/ref:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())
		CheckComponentInArchive(env, ldesc, OUT)
	})

	It("transfers ctf with --closure --lookup", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", "--recursive", "--lookup", ARCH, ARCH2, ARCH2, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test2:v1"...
  transferring version "github.com/mandelsoft/test:v1"...
  ...resource 0 testdata[plainText]...
  ...resource 1 value[ociImage](ocm/value:v2.0)...
  ...resource 2 ref[ociImage](ocm/ref:v2.0)...
  ...adding component version...
...resource 0 otherdate[plainText]...
...adding component version...
2 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())
		tgt, err := ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, OUT, 0, accessio.PathFileSystem(env.FileSystem()))
		Expect(err).To(Succeed())
		defer Close(tgt, "ctf")

		list, err := tgt.ComponentLister().GetComponents("", true)
		Expect(err).To(Succeed())
		Expect(list).To(ContainElements([]string{COMPONENT2, COMPONENT}))

		Expect(tgt.ExistsComponentVersion(COMPONENT2, VERSION)).To(BeTrue())

		CheckComponent(env, ldesc, tgt)
	})

	It("transfers ctf creating bom file", func() {
		BOM := "/tmp/bom.json"
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--bom-file="+BOM, "--copy-resources", "--recursive", "--lookup", ARCH, ARCH2, ARCH2, OUT)).To(Succeed())

		Expect(env.FileExists(BOM)).To(BeTrue())

		data := Must(env.ReadFile(BOM))
		Expect(data).To(YAMLEqual(`
  componentVersions:
  - component: github.com/mandelsoft/test
    version: v1
  - component: github.com/mandelsoft/test2
    version: v1
`))
	})

	It("transfers ctf to tgz with type option", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", "--type", accessio.FormatTGZ.String(), ARCH, ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[plainText]...
...resource 1 value[ociImage](ocm/value:v2.0)...
...resource 2 ref[ociImage](ocm/ref:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		CheckComponentInArchive(env, ldesc, OUT)
	})

	It("transfers ctf to tgz", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", ARCH, ARCH, accessio.FormatTGZ.String()+"::"+OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[plainText]...
...resource 1 value[ociImage](ocm/value:v2.0)...
...resource 2 ref[ociImage](ocm/ref:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		CheckComponentInArchive(env, ldesc, OUT)
	})

	It("transfers ctf to ctf+tgz", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", ARCH, ARCH, "ctf+"+accessio.FormatTGZ.String()+"::"+OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[plainText]...
...resource 1 value[ociImage](ocm/value:v2.0)...
...resource 2 ref[ociImage](ocm/ref:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		CheckComponentInArchive(env, ldesc, OUT)
	})

	It("transfers ctf to ctf+tgz with config option", func() {
		cfg := handlercfg.NewConfig()
		cfg.ResourcesByValue = utils.BoolP(true)

		mod := func(ctx clictx.Context, cmd *cobra.Command) {
			if cmd == nil {
				MustBeSuccessful(ctx.ConfigContext().ApplyConfig(cfg, "explicit"))
			}
		}
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(mod, "transfer", "components", ARCH, ARCH, "ctf+"+accessio.FormatTGZ.String()+"::"+OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[plainText]...
...resource 1 value[ociImage](ocm/value:v2.0)...
...resource 2 ref[ociImage](ocm/ref:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		CheckComponentInArchive(env, ldesc, OUT)
	})

	It("transfers ctf to ctf+tgz with config set", func() {
		cfg := handlercfg.NewConfig()
		cfg.ResourcesByValue = utils.BoolP(true)

		cfgcfg := config.New()
		cfgcfg.AddSet("transfer", "standard transfer options to use")
		cfgcfg.AddConfigToSet("transfer", cfg)

		mod := func(ctx clictx.Context, cmd *cobra.Command) {
			if cmd == nil {
				MustBeSuccessful(ctx.ConfigContext().ApplyConfig(cfgcfg, "explicit"))
			}
		}
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(mod, "--config-set", "transfer", "transfer", "components", ARCH, ARCH, "ctf+"+accessio.FormatTGZ.String()+"::"+OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[plainText]...
...resource 1 value[ociImage](ocm/value:v2.0)...
...resource 2 ref[ociImage](ocm/ref:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.FileExists(OUT)).To(BeTrue())
		CheckComponentInArchive(env, ldesc, OUT)
	})
})
