//go:build unix

package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/testutils"
	. "github.com/open-component-model/ocm/pkg/env/builder"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
)

const (
	ARCH     = "ctf"
	COMP     = "github.com/mandelsoft/comp"
	VERS     = "1.0.0"
	PROVIDER = "mandelsoft"
)

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var env *Builder
	var plugins TempPluginDir

	var accessSpec ocm.AccessSpec

	BeforeEach(func() {
		var err error

		accessSpec, err = ocm.NewGenericAccessSpec(`
type: test
someattr: value
`)
		Expect(err).To(Succeed())

		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugins, registry = Must2(ConfigureTestPlugins2(env, "testdata"))
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("registers access methods", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERS, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", VERS, "PlainText", metav1.ExternalRelation, func() {
						env.Access(accessSpec)
					})
				})
			})
		})

		repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo)

		cv := Must(repo.LookupComponentVersion(COMP, VERS))
		defer Close(cv)

		r := Must(cv.GetResourceByIndex(0))

		m := Must(r.AccessMethod())
		Expect(m.MimeType()).To(Equal("plain/text"))

		data := Must(m.Get())
		Expect(string(data)).To(Equal("test content\n{\"someattr\":\"value\",\"type\":\"test\"}\n"))
	})
})
