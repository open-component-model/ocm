//go:build unix

package accessmethods_test

import (
	"fmt"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/api/ocm/plugin/testutils"

	"github.com/mandelsoft/vfs/pkg/utils"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin/registration"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/demoplugin/accessmethods"
)

const (
	CONTENT = "testdata"
)

var _ = Describe("demoplugin", func() {
	Context("lib", func() {
		var env *Builder
		var plugins TempPluginDir
		var osf *os.File

		BeforeEach(func() {
			env = NewBuilder(TestData())
			plugins = Must(ConfigureTestPlugins(env, "testdata"))

			registry := plugincacheattr.Get(env)
			Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get("demo")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))

			MustBeSuccessful(env.FileSystem().MkdirAll("data", 0o700))

			f := Must(env.FileSystem().Create("/data/test"))
			Must(f.Write([]byte(CONTENT)))

			osf = utils.OSFile(f)
			Expect(osf).NotTo(BeNil())
		})

		AfterEach(func() {
			plugins.Cleanup()
			env.Cleanup()
		})

		It("get templ file", func() {
			p := osf.Name()[len(os.TempDir())+1:]
			fmt.Printf("%s\n", p)

			spec := accessmethods.AccessSpec{
				ObjectVersionedType: runtime.NewVersionedTypedObject(accessmethods.NAME),
				Path:                p,
				MediaType:           mime.MIME_TEXT,
			}
			a := Must(env.OCMContext().AccessSpecForSpec(spec))

			cv := &cpi.DummyComponentVersionAccess{Context: env.OCMContext()}
			hints := cpi.ReferenceHint(a, cv)
			Expect(len(hints)).To(Equal(1))
			Expect(hints[0].GetReference()).To(Equal(p))
			Expect(a.Describe(env.OCMContext())).To(Equal("temp file " + p))

			m := Must(a.AccessMethod(cv))
			data := Must(m.Get())

			Expect(string(data)).To(Equal(CONTENT))
		})
	})
})
