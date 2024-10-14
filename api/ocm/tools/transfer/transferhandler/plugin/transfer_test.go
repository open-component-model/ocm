package plugin_test

import (
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/oci"
	. "ocm.software/ocm/api/oci/testhelper"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
)

const OCIPATH1 = "oci1"
const OCIPATH2 = "oci2"

const OCIHOST1 = "source"
const OCIHOST2 = "target"

const COMPONENT = "acme.org/component"
const VERSION = "1.0.0"
const PROVIDER = "acme.org"

var _ = Describe("transport with plugin based transfer handler", func() {
	var env *Builder
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewBuilder(TestData())
		plugins = Must(ConfigureTestPlugins(env, "testdata/plugins"))

		env.OCICommonTransport(OCIPATH1, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})

		env.OCICommonTransport(OCIPATH2, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})

		FakeOCIRepo(env, OCIPATH1, OCIHOST1)
		FakeOCIRepo(env, OCIPATH2, OCIHOST2)

		env.OCMCommonTransport(OCIPATH1, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("image", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST1+".alias", OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("loads plugin", func() {
		registry := plugincacheattr.Get(env)
		//	Expect(registration.RegisterExtensions(env)).To(Succeed())
		p := registry.Get(PLUGIN)
		Expect(p).NotTo(BeNil())
		Expect(p.Error()).To(Equal(""))
	})

	DescribeTable("transfers per handler", func(cfgfile string) {
		config := Must(os.ReadFile("testdata/" + cfgfile))
		p, buf := common.NewBufferedPrinter()
		topts := append([]transferhandler.TransferOption{
			transfer.WithPrinter(p),
			transferhandler.WithConfig(config),
		})

		h := Must(transferhandler.For(env).ByName(env, "plugin/transferplugin/demo", topts...))

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, OCIPATH1, 0, env))
		defer Close(src, "src")

		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "cv")

		tgt := Must(ctf.Open(env, accessobj.ACC_WRITABLE, OCIPATH2, 0, env))
		ctgt := accessio.OnceCloser(tgt)
		defer Close(ctgt, "tgt")

		MustBeSuccessful(transfer.TransferWithHandler(p, cv, tgt, h))

		options := &standard.Options{}
		transferhandler.ApplyOptions(options, topts...)

		out := `
  transferring version "acme.org/component:1.0.0"...
  ...resource 0 image\[ociImage\]\(ocm/value:v2.0\)...
  ...adding component version...
`
		if cfgfile == "config" {
			out = `
  transferring version "acme.org/component:1.0.0"...
  ...adding component version...
`
		}
		Expect(string(buf.Bytes())).To(StringMatchTrimmedWithContext(utils.Crop(out, 2)))
		MustBeSuccessful(ctgt.Close())

		tgt = Must(ctf.Open(env, accessobj.ACC_READONLY, OCIPATH2, 0, env))
		ctgt = accessio.OnceCloser(tgt)
		defer Close(ctgt, "tgt2")

		tcv := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		ctcv := accessio.OnceCloser(tcv)
		defer Close(ctcv, "tcv")

		r := Must(tcv.GetResourceByIndex(0))
		acc := Must(r.Access())

		atype := localblob.Type
		if cfgfile == "config" {
			atype = ociartifact.Type
		}
		Expect(acc.GetKind()).To(Equal(atype))

		info := acc.Info(env.OCMContext())

		infotxt := "sha256:" + H_OCIARCHMANIFEST1
		if cfgfile == "config" {
			infotxt = "ocm/value:v2.0"
		}
		Expect(info.Info).To(Equal(infotxt))

		MustBeSuccessful(ctcv.Close())
		MustBeSuccessful(ctgt.Close())
	},
		Entry("with matching oci host", "transferconfig"),
		Entry("without matching oci host", "config"),
	)
})
