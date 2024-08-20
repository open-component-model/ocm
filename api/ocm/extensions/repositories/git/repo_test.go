package git_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-git/go-billy/v5"
	gitgo "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/filesystem"
	. "github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tonglil/buflogr"
	"ocm.software/ocm/api/oci/artdesc"
	gitrepo "ocm.software/ocm/api/oci/extensions/repositories/git"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
	"ocm.software/ocm/api/ocm/extensions/repositories/git"
	. "ocm.software/ocm/api/ocm/testhelper"
	techgit "ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/refmgmt"
)

const (
	COMPONENT   = "ocm.software/ocm"
	VERSION     = "1.0.0"
	REMOTE_REPO = "repo.git"
)

var _ = Describe("access method", func() {

	// remoteFS contains the remote repository on the filesystem
	// pathFS contains the local PWD
	// repFS contains the local representation of the repository, meaning the cloned repo to work on the Repository
	var remoteFS, pathFS, repFS vfs.FileSystem
	// access contains the access configuration to the above filesystems
	var access accessio.Options

	// repoURL is the URL specification to access the remote repository in remoteFS
	var repoURL string

	// remoteRepo is the remote repository that can be used for test assertions on pushed content
	var remoteRepo *gitgo.Repository

	var opts git.Options

	ctx := ocm.DefaultContext()

	BeforeEach(func() {
		By("setting up test filesystems")
		basePath := GinkgoT().TempDir()
		baseFS := Must(cwdfs.New(osfs.New(), basePath))
		for _, dir := range []string{"remote", "path", "rep"} {
			Expect(os.Mkdir(filepath.Join(basePath, dir), 0777)).To(Succeed())
		}
		remoteFS = Must(projectionfs.New(baseFS, "remote"))
		pathFS = Must(projectionfs.New(baseFS, "path"))
		repFS = Must(projectionfs.New(baseFS, "rep"))

		access = &accessio.StandardOptions{
			PathFileSystem: pathFS,
			Representation: repFS,
		}
	})

	AfterEach(func() {
		Expect(Must(vfs.ReadDir(pathFS, "."))).To(BeEmpty(), "nothing of the CTF should be stored in the path, "+
			"because everything should be handled in the representation which contains the local git repository")
	})

	BeforeEach(func() {
		By("setting up local bare git repository to work against when pushing/updating")
		billy := techgit.VFSBillyFS(remoteFS)
		client.InstallProtocol("file", server.NewClient(server.NewFilesystemLoader(billy)))
		remoteRepo = Must(newBareTestRepo(billy, REMOTE_REPO, gitgo.InitOptions{}))
		// now that we have a bare repository, we can reference it via URL to access it like a remote repository
		repoURL = fmt.Sprintf("file:///%s", REMOTE_REPO)
	})

	BeforeEach(func() {
		opts = git.Options{
			Author: &git.Author{
				Name:  fmt.Sprintf("OCM Test Case: %s", GinkgoT().Name()),
				Email: "dummy@ocm.software",
			},
			Options: access,
		}
	})

	It("adds naked component version and later lookup", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(git.Create(ctx, git.ACC_WRITABLE|git.ACC_CREATE, repoURL, opts))
		final.Close(a, "repository")
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c, "component")

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv, "version")

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		componentCommitExpectation := gitrepo.GenerateCommitMessageForNamespace(gitrepo.OperationUpdate, fmt.Sprintf("component-descriptors/%s", COMPONENT))
		descriptorCommitExpectation := regexp.MustCompile(fmt.Sprintf("%s: %s blob.* of type %s",
			regexp.QuoteMeta(gitrepo.CommitPrefix),
			gitrepo.OperationAdd,
			regexp.QuoteMeta(componentmapping.ComponentDescriptorTarMimeType)),
		)
		descriptorConfigCommitExpectation := regexp.MustCompile(fmt.Sprintf("%s: %s blob.* of type %s",
			regexp.QuoteMeta(gitrepo.CommitPrefix),
			gitrepo.OperationAdd,
			regexp.QuoteMeta(componentmapping.ComponentDescriptorConfigMimeType)),
		)
		manifestAddCommitExpectation := regexp.MustCompile(fmt.Sprintf("%s: %s artifact .* %s",
			regexp.QuoteMeta(gitrepo.CommitPrefix),
			gitrepo.OperationAdd,
			regexp.QuoteMeta(fmt.Sprintf("(%s)", artdesc.MediaTypeImageManifest))),
		)
		manifestUpdateCommitExpectation := regexp.MustCompile(fmt.Sprintf("%s: %s manifest .* %s",
			regexp.QuoteMeta(gitrepo.CommitPrefix),
			gitrepo.OperationUpdate,
			regexp.QuoteMeta(fmt.Sprintf("(%s)", artdesc.MediaTypeImageManifest))),
		)

		componentUpdate := 0
		descriptorCommits := 0
		manifestUpdateCommits := 0
		commits := Must(remoteRepo.CommitObjects())
		Expect(commits.ForEach(func(commit *object.Commit) error {
			Expect(commit.Author.Name).To(Equal(opts.Author.Name))
			Expect(commit.Author.Email).To(Equal(opts.Author.Email))

			if commit.Message == componentCommitExpectation {
				componentUpdate++
			} else if descriptorCommitExpectation.MatchString(commit.Message) || descriptorConfigCommitExpectation.MatchString(commit.Message) {
				descriptorCommits++
			} else if manifestUpdateCommitExpectation.MatchString(commit.Message) || manifestAddCommitExpectation.MatchString(commit.Message) {
				manifestUpdateCommits++
			}
			return nil
		})).To(Succeed())
		Expect(componentUpdate).To(Equal(1))
		Expect(descriptorCommits).To(Equal(2))
		Expect(manifestUpdateCommits).To(Equal(2))

		refmgmt.AllocLog.Trace("opening ctf")
		a = Must(git.Open(ctx, git.ACC_READONLY, repoURL, opts))
		final.Close(a)

		refmgmt.AllocLog.Trace("lookup component")
		c, err := a.LookupComponent(COMPONENT)
		Expect(err).ToNot(HaveOccurred())
		final.Close(c)

		refmgmt.AllocLog.Trace("lookup version")
		cv = Must(c.LookupVersion(VERSION))
		final.Close(cv)

		refmgmt.AllocLog.Trace("closing")
		MustBeSuccessful(final.Finalize())
	})

	It("adds naked component version and later shortcut lookup", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(git.Create(ctx, git.ACC_WRITABLE|git.ACC_CREATE, repoURL, opts))
		final.Close(a, "repository")
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c, "component")

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv, "version")

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		refmgmt.AllocLog.Trace("opening ctf")
		a = Must(git.Open(ctx, git.ACC_READONLY, repoURL, opts))
		final.Close(a)

		refmgmt.AllocLog.Trace("lookup component version")
		cv = Must(a.LookupComponentVersion(COMPONENT, VERSION))
		final.Close(cv)

		refmgmt.AllocLog.Trace("closing")
		MustBeSuccessful(final.Finalize())
	})

	It("adds component version", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(git.Create(ctx, git.ACC_WRITABLE|git.ACC_CREATE, repoURL, opts))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		// add resource
		MustBeSuccessful(cv.SetResourceBlob(compdesc.NewResourceMeta("text1", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
		Expect(Must(cv.GetResource(compdesc.NewIdentity("text1"))).Meta().Digest).To(Equal(DS_TESTDATA))

		// add resource with digest
		meta := compdesc.NewResourceMeta("text2", resourcetypes.PLAIN_TEXT, metav1.LocalRelation)
		meta.SetDigest(DS_TESTDATA)
		MustBeSuccessful(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
		Expect(Must(cv.GetResource(compdesc.NewIdentity("text2"))).Meta().Digest).To(Equal(DS_TESTDATA))

		// reject resource with wrong digest
		meta = compdesc.NewResourceMeta("text3", resourcetypes.PLAIN_TEXT, metav1.LocalRelation)
		meta.SetDigest(TextResourceDigestSpec("fake"))
		Expect(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil)).To(MatchError("unable to set resource: digest mismatch: " + D_TESTDATA + " != fake"))

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		a = Must(git.Open(ctx, git.ACC_READONLY, repoURL, opts))
		final.Close(a)

		cv = Must(a.LookupComponentVersion(COMPONENT, VERSION))
		final.Close(cv)
	})

	It("adds omits unadded new component version", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(git.Create(ctx, git.ACC_WRITABLE|git.ACC_CREATE, repoURL, opts))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		MustBeSuccessful(final.Finalize())

		a = Must(git.Open(ctx, git.ACC_READONLY, repoURL, opts))
		final.Close(a)

		_, err := a.LookupComponentVersion(COMPONENT, VERSION)

		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("component version \"%[1]s:%[2]s\" not found: oci artifact \"%[2]s\" not found in component-descriptors/%[1]s", COMPONENT, VERSION))))
	})

	It("provides error for invalid bloc access", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(git.Create(ctx, git.ACC_WRITABLE|git.ACC_CREATE, repoURL, opts))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		// add resource
		Expect(ErrorFrom(cv.SetResourceBlob(compdesc.NewResourceMeta("text1", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blobaccess.ForFile(mime.MIME_TEXT, "non-existing-file"), "", nil))).To(MatchError(`file "non-existing-file" not found`))

		MustBeSuccessful(final.Finalize())
	})

	It("logs diff", func() {
		MustBeSuccessful(accessio.FormatDirectory.ApplyOption(opts.Options))
		r := Must(git.Open(ctx, git.ACC_CREATE, repoURL, opts))
		defer Close(r, "repo")

		c := Must(r.LookupComponent("acme.org/test"))
		defer Close(c, "comp")

		cv := Must(c.NewVersion("v1"))

		ocmlog.PushContext(nil)
		ocmlog.Context().AddRule(logging.NewConditionRule(logging.DebugLevel, genericocireg.TAG_CDDIFF))
		var buf bytes.Buffer
		def := buflogr.NewWithBuffer(&buf)
		ocmlog.Context().SetBaseLogger(def)
		defer ocmlog.Context().ResetRules()
		defer ocmlog.PopContext()

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(cv.Close())

		cv = Must(c.LookupVersion("v1"))
		cv.GetDescriptor().Provider.Name = "acme.org"
		MustBeSuccessful(cv.Close())
		Expect("\n" + buf.String()).To(Equal(fmt.Sprintf(`
V[4] component descriptor has been changed realm ocm realm ocm/oci/mapping diff [ComponentSpec.ObjectMeta.Provider.Name: acme != %[1]s]
V[4] component descriptor has been changed realm ocm realm ocm/oci/mapping diff [ComponentSpec.ObjectMeta.Provider.Name: acme != %[1]s]
`, cv.GetDescriptor().Provider.Name)))
	})

	It("handles readonly mode", func() {
		MustBeSuccessful(accessio.FormatDirectory.ApplyOption(opts.Options))
		r := Must(git.Open(ctx, git.ACC_CREATE, repoURL, opts))
		defer Close(r, "repo")

		c := Must(r.LookupComponent("acme.org/test"))
		defer Close(c, "comp")

		cv := Must(c.NewVersion("v1"))

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(cv.Close())

		cv = Must(c.LookupVersion("v1"))
		cv.SetReadOnly()
		Expect(cv.IsReadOnly()).To(BeTrue())
		cv.GetDescriptor().Provider.Name = "acme.org"
		ExpectError(cv.Close()).To(MatchError(accessio.ErrReadOnly))
	})

	It("handles readonly mode on repo", func() {
		MustBeSuccessful(accessio.FormatDirectory.ApplyOption(opts.Options))
		r := Must(git.Open(ctx, git.ACC_CREATE, repoURL, opts))
		defer Close(r, "repo")

		c := Must(r.LookupComponent("acme.org/test"))
		defer Close(c, "comp")

		cv := Must(c.NewVersion("v1"))

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(cv.Close())

		r.SetReadOnly()
		cv = Must(c.LookupVersion("v1"))
		Expect(cv.IsReadOnly()).To(BeTrue())
		cv.GetDescriptor().Provider.Name = "acme.org"
		ExpectError(cv.Close()).To(MatchError(accessio.ErrReadOnly))

		ExpectError(c.NewVersion("v2")).To(MatchError(accessio.ErrReadOnly))
	})
})

func newBareTestRepo(fs billy.Filesystem, path string, opts gitgo.InitOptions) (*gitgo.Repository, error) {
	var wt, dot billy.Filesystem

	var err error
	dot, err = fs.Chroot(path)
	if err != nil {
		return nil, err
	}
	wt, err = fs.Chroot(path)
	if err != nil {
		return nil, err
	}

	s := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	r, err := gitgo.InitWithOptions(s, wt, opts)
	if err != nil {
		return nil, err
	}

	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}

	err = r.Storer.SetConfig(cfg)
	if err != nil {
		return nil, err
	}

	return r, err
}
