package git_test

import (
	"embed"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mandelsoft/filepath/pkg/filepath"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	self "ocm.software/ocm/api/tech/git"
)

//go:embed testdata/repo
var testData embed.FS

var _ = Describe("standard tests with local file repo", func() {
	var (
		ctx                 ocm.Context
		expectedBlobContent []byte
	)

	ctx = ocm.New()

	BeforeEach(func() {
		tempVFS, err := cwdfs.New(osfs.New(), GinkgoT().TempDir())
		Expect(err).ToNot(HaveOccurred())
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: ".", Filesystem: tempVFS})
		vfsattr.Set(ctx, tempVFS)
	})

	var repoDir string
	var repoURL string
	var ref string
	var commit string

	BeforeEach(func() {
		repoDir = GinkgoT().TempDir() + filepath.PathSeparatorString + "repo"

		repo := Must(git.PlainInit(repoDir, false))

		repoBase := filepath.Join("testdata", "repo")
		repoTestData := Must(testData.ReadDir(repoBase))

		for _, entry := range repoTestData {
			path := filepath.Join(repoBase, entry.Name())
			repoPath := filepath.Join(repoDir, entry.Name())

			file := Must(testData.Open(path))

			fileInRepo := Must(os.OpenFile(
				repoPath,
				os.O_CREATE|os.O_RDWR|os.O_TRUNC,
				0o600,
			))

			Must(io.Copy(fileInRepo, file))

			Expect(fileInRepo.Close()).To(Succeed())
			Expect(file.Close()).To(Succeed())
		}

		wt := Must(repo.Worktree())
		Expect(wt.AddGlob("*")).To(Succeed())
		commit = Must(wt.Commit("OCM Test Commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "OCM Test",
				Email: "dummy@ocm.software",
				When:  time.Now(),
			},
		})).String()

		path := filepath.Join("testdata", "repo", "file_in_repo")
		repoURL = fmt.Sprintf("file://%s", repoDir)
		ref = plumbing.Master.String()

		expectedBlobContent = Must(testData.ReadFile(path))
	})

	It("Resolver client can setup repository", func(ctx SpecContext) {
		client := Must(self.NewClient(self.ClientOptions{
			URL:    repoURL,
			Ref:    ref,
			Commit: commit,
		}))

		tempVFS, err := projectionfs.New(osfs.New(), GinkgoT().TempDir())
		Expect(err).ToNot(HaveOccurred())

		Expect(client.Setup(ctx, tempVFS)).To(Succeed())

		repo := Must(client.Repository(ctx))
		Expect(repo).ToNot(BeNil())

		file := Must(tempVFS.Stat("file_in_repo"))
		Expect(file.Size()).To(Equal(int64(len(expectedBlobContent))))
	})
})
