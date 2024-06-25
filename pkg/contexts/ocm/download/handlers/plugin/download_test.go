//go:build unix

package plugin_test

import (
	"fmt"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/testutils"
	. "github.com/open-component-model/ocm/pkg/env/builder"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const PLUGIN = "test"

const (
	ARCH     = "ctf"
	COMP     = "github.com/mandelsoft/comp"
	VERS     = "1.0.0"
	PROVIDER = "mandelsoft"
	RSCTYPE  = "TestArtifact"
	MEDIA    = "text/plain"
)

const (
	REPOTYPE = "test/v1"
	ACCTYPE  = "test/v1"
	CONTENT  = "some test content\n"
	HINT     = "given"
)

type AccessSpec struct {
	runtime.ObjectVersionedType
	Path       string `json:"path"`
	MediaType  string `json:"mediaType"`
	Repository string `json:"repo"`
}

func NewAccessSpec(media, path, repo string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.ObjectVersionedType{Type: ACCTYPE},
		MediaType:           media,
		Path:                path,
		Repository:          repo,
	}
}

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var repodir string
	var env *Builder
	var plugins TempPluginDir

	BeforeEach(func() {
		repodir = Must(os.MkdirTemp(os.TempDir(), "uploadtest-*"))

		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugins, registry = Must2(ConfigureTestPlugins2(env, "testdata"))
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())

		ctx.ConfigContext().ApplyConfig(config.New(PLUGIN, []byte(fmt.Sprintf(`{"root": "`+repodir+`"}`))), "plugin config")
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
		os.RemoveAll(repodir)
	})

	It("downloads resource", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERS, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", VERS, RSCTYPE, metav1.LocalRelation, func() {
						env.Hint(HINT)
						env.BlobStringData(MEDIA, CONTENT)
					})
				})
			})
		})

		repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "source repo")

		cv := Must(repo.LookupComponentVersion(COMP, VERS))
		defer Close(cv, "source version")

		MustFailWithMessage(plugin.RegisterDownloadHandler(env.OCMContext(), "test", "", nil, download.ForArtifactType("blah")), "no downloader found for [art:\"blah\", media:\"\"]")
		MustBeSuccessful(plugin.RegisterDownloadHandler(env.OCMContext(), "test", "", nil, download.ForArtifactType(RSCTYPE)))

		racc := Must(cv.GetResourceByIndex(0))

		file := vfs.Join(env.FileSystem(), repodir, "download")

		octx, buf := out.NewBuffered()
		ok, eff, err := download.For(env).Download(common.NewPrinter(octx.StdOut()), racc, file, nil)

		MustBeSuccessful(err)
		Expect(buf.String()).To(Equal(""))
		Expect(eff).To(Equal(file))
		Expect(ok).To(BeTrue())

		data := Must(os.ReadFile(file))
		Expect(string(data)).To(StringEqualTrimmedWithContext(`
downloaded
` + CONTENT))
	})
})
