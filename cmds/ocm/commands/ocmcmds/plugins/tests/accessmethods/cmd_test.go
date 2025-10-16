package accessmethods_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/plugin/plugins"
	"ocm.software/ocm/api/ocm/plugin/registration"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	CA      = "/tmp/ca"
	VERSION = "v1"
)

var _ = Describe("Add with new access method", func() {
	var env *TestEnv
	var ctx ocm.Context
	var registry plugins.Set
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewTestEnv(TestData())
		ctx = env.OCMContext()
		plugins, registry = Must2(ConfigureTestPlugins2(env, "testdata"))

		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())

		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", CA)).To(Succeed())
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("adds external resource by options", func() {
		Expect(env.Execute("add", "resources", CA,
			"--type", "testContent",
			"--name", "text",
			"--version", "v0.1.0",
			"--accessType", "test",
			"--accessPath", "textfile",
			"--mediaType", "text/plain")).To(Succeed())
		data := Must(env.ReadFile(env.Join(CA, comparch.ComponentDescriptorFileName)))
		cd := Must(compdesc.Decode(data))
		Expect(len(cd.Resources)).To(Equal(1))

		r := Must(cd.GetResourceByIdentity(metav1.NewIdentity("text")))
		Expect(r.Type).To(Equal("testContent"))
		Expect(r.Version).To(Equal("v0.1.0"))
		Expect(r.Relation).To(Equal(metav1.ResourceRelation("external")))

		Expect(r.Access.GetType()).To(Equal("test"))
		acc := Must(env.OCMContext().AccessSpecForSpec(r.Access))
		var myacc AccessSpec

		MustBeSuccessful(json.Unmarshal(Must(json.Marshal(acc)), &myacc))
		Expect(myacc).To(Equal(AccessSpec{Type: "test", Path: "textfile", MediaType: "text/plain"}))

		m := Must(acc.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()}))
		data = Must(m.Get())
		Expect(string(data)).To(Equal("test content\n{\"mediaType\":\"text/plain\",\"path\":\"textfile\",\"type\":\"test\"}\n"))
	})
})
