//go:build unix

package uploaders_test

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/api/ocm/plugin/testutils"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/utils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/plugin"
	"ocm.software/ocm/api/ocm/plugin/registration"
	"ocm.software/ocm/api/ocm/refhints"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/demoplugin/accessmethods"
	"ocm.software/ocm/cmds/demoplugin/uploaders"
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

			registry := plugincacheattr.Get(env)

			_, comps, _ := vfs.SplitPath(env.FileSystem(), p)

			tgt := vfs.Join(env.FileSystem(), comps[0], "upload")
			cfg := uploaders.NewTarget(tgt)
			raw := Must(json.Marshal(cfg))

			uploader := Must(plugin.New(registry.Get("demo"), "demo", raw))
			Expect(uploader).NotTo(BeNil())

			sctx := &cpi.DummyStorageContext{
				Context: env.OCMContext(),
			}
			cv := &cpi.DummyComponentVersionAccess{
				Context: env.OCMContext(),
			}

			hint := "uploaded"
			blob := blobaccess.ForString(mime.MIME_TEXT, CONTENT)
			spec := Must(uploader.StoreBlob(blob, resourcetypes.PLAIN_TEXT, refhints.DefaultList(accessmethods.ReferenceHint, hint), nil, sctx))

			Expect(spec).NotTo(BeNil())

			hints := cpi.ReferenceHint(spec, cv)
			Expect(len(hints)).To(Equal(1))
			name := vfs.Join(env.FileSystem(), tgt, hint)
			Expect(hints[0].GetReference()).To(Equal(name))

			f := filepath.Join(os.TempDir(), tgt, hint)
			data := Must(os.ReadFile(f))
			fmt.Printf("written %q: %s\n", f, string(data))
			Expect(string(data)).To(Equal(CONTENT))
		})
	})
})
