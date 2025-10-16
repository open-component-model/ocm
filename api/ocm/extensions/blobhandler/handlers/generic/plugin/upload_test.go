//go:build unix

package plugin_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/blobhandler"
	blobplugin "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/plugin"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/config"
	"ocm.software/ocm/api/ocm/plugin/plugins"
	"ocm.software/ocm/api/ocm/plugin/registration"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/runtime"
)

const PLUGIN = "test"

const (
	ARCH     = "ctf"
	OUT      = "/tmp/res"
	COMP     = "github.com/mandelsoft/comp"
	VERS     = "1.0.0"
	PROVIDER = "mandelsoft"
	RSCTYPE  = "TestArtifact"
	MEDIA    = "text/plain"
)

const (
	REPOTYPE = "test/v1"
	ACCTYPE  = "test/v1"
	REPO     = "plugin"
	CONTENT  = "some test content\n"
	HINT     = "given"
)

type RepoSpec struct {
	runtime.ObjectVersionedType
	Path string `json:"path"`
}

func NewRepoSpec(path string) *RepoSpec {
	return &RepoSpec{
		ObjectVersionedType: runtime.ObjectVersionedType{Type: REPOTYPE},
		Path:                path,
	}
}

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

	accessSpec := NewAccessSpec(MEDIA, "given", REPO)
	repoSpec := NewRepoSpec(REPO)

	BeforeEach(func() {
		repodir = Must(os.MkdirTemp(os.TempDir(), "uploadtest-*"))

		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugins, registry = Must2(ConfigureTestPlugins2(env, "testdata"))
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())

		ctx.ConfigContext().ApplyConfig(config.New(PLUGIN, []byte(fmt.Sprintf(`{"root": "`+repodir+`"}`))), "plugin config")
		registration.RegisterExtensions(ctx)

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERS, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", VERS, RSCTYPE, metav1.LocalRelation, func() {
						env.Hint(HINT)
						env.BlobStringData(MEDIA, CONTENT)
						// env.Access(NewAccessSpec(MEDIA, "given", "dummy"))
					})
				})
			})
		})
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
		os.RemoveAll(repodir)
	})

	It("uploads artifact", func() {
		repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "source repo")

		cv := Must(repo.LookupComponentVersion(COMP, VERS))
		defer Close(cv, "source version")

		_, _, err := blobplugin.RegisterBlobHandler(env.OCMContext(), "test", "", RSCTYPE, "", []byte("{}"))
		fmt.Printf("error %q\n", err)
		MustFailWithMessage(err, "plugin uploader test/testuploader: error processing plugin command upload: path missing in repository spec")
		repospec := Must(json.Marshal(repoSpec))
		name, keys, err := blobplugin.RegisterBlobHandler(env.OCMContext(), "test", "", RSCTYPE, "", repospec)
		MustBeSuccessful(err)
		Expect(name).To(Equal("testuploader"))
		Expect(keys).To(Equal(plugin.UploaderKeySet{}.Add(plugin.UploaderKey{}.SetArtifact(RSCTYPE, ""))))

		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer Close(tgt, "target repo")

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, Must(standard.New(standard.ResourcesByValue()))))
		Expect(env.DirExists(OUT)).To(BeTrue())

		Expect(vfs.FileExists(osfs.New(), filepath.Join(repodir, REPO, HINT))).To(BeTrue())

		tcv := Must(tgt.LookupComponentVersion(COMP, VERS))
		defer Close(tcv, "target version")

		r := Must(tcv.GetResourceByIndex(0))
		a := Must(r.Access())

		var spec AccessSpec
		MustBeSuccessful(json.Unmarshal(Must(json.Marshal(a)), &spec))
		Expect(spec).To(Equal(*accessSpec))

		m := Must(a.AccessMethod(tcv))
		defer Close(m, "method")

		Expect(string(Must(m.Get()))).To(Equal(CONTENT))
	})

	It("uploads after abstract registration", func() {
		repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "source repo")

		cv := Must(repo.LookupComponentVersion(COMP, VERS))
		defer Close(cv, "source version")

		MustFailWithMessage(blobhandler.RegisterHandlerByName(ctx, "plugin/test", []byte("{}"), blobhandler.ForArtifactType(RSCTYPE)),
			"plugin uploader test/testuploader: error processing plugin command upload: path missing in repository spec")
		repospec := Must(json.Marshal(repoSpec))
		MustBeSuccessful(blobhandler.RegisterHandlerByName(ctx, "plugin/test", repospec))

		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer Close(tgt, "target repo")

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, Must(standard.New(standard.ResourcesByValue()))))
		Expect(env.DirExists(OUT)).To(BeTrue())

		Expect(vfs.FileExists(osfs.New(), filepath.Join(repodir, REPO, HINT))).To(BeTrue())

		tcv := Must(tgt.LookupComponentVersion(COMP, VERS))
		defer Close(tcv, "target version")

		r := Must(tcv.GetResourceByIndex(0))
		a := Must(r.Access())

		var spec AccessSpec
		MustBeSuccessful(json.Unmarshal(Must(json.Marshal(a)), &spec))
		Expect(spec).To(Equal(*accessSpec))

		m := Must(a.AccessMethod(tcv))
		defer Close(m, "method")

		Expect(string(Must(m.Get()))).To(Equal(CONTENT))
	})
})
