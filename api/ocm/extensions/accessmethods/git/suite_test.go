package git_test

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	me "ocm.software/ocm/api/ocm/extensions/accessmethods/git"
)

//go:embed testdata/repo
var testData embed.FS

var _ = Describe("Method", func() {
	var (
		ctx                 ocm.Context
		expectedBlobContent []byte
		accessSpec          *me.AccessSpec
	)

	ctx = ocm.New()

	BeforeEach(func() {
		tempVFS, err := cwdfs.New(osfs.New(), GinkgoT().TempDir())
		Expect(err).ToNot(HaveOccurred())
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: ".", Filesystem: tempVFS})
		vfsattr.Set(ctx, tempVFS)
	})

	BeforeEach(func() {
		repoDir := GinkgoT().TempDir() + filepath.PathSeparatorString + "repo"

		repo, err := git.PlainInit(repoDir, false)
		Expect(err).ToNot(HaveOccurred())

		repoBase := filepath.Join("testdata", "repo")
		repoTestData, err := testData.ReadDir(repoBase)
		Expect(err).ToNot(HaveOccurred())

		for _, entry := range repoTestData {
			path := filepath.Join(repoBase, entry.Name())
			repoPath := filepath.Join(repoDir, entry.Name())

			file, err := testData.Open(path)
			Expect(err).ToNot(HaveOccurred())

			fileInRepo, err := os.OpenFile(
				repoPath,
				os.O_CREATE|os.O_RDWR|os.O_TRUNC,
				0600,
			)
			Expect(err).ToNot(HaveOccurred())

			_, err = io.Copy(fileInRepo, file)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileInRepo.Close()).To(Succeed())
			Expect(file.Close()).To(Succeed())
		}

		wt, err := repo.Worktree()
		Expect(err).ToNot(HaveOccurred())
		Expect(wt.AddGlob("*")).To(Succeed())
		_, err = wt.Commit("OCM Test Commit", &git.CommitOptions{})
		Expect(err).ToNot(HaveOccurred())

		accessSpec = me.New(
			fmt.Sprintf("file://%s", repoDir),
			string(plumbing.Master),
			".",
		)
	})

	It("downloads artifacts", func() {
		m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
		Expect(err).ToNot(HaveOccurred())
		content, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(content).To(Equal(expectedBlobContent))
	})

})
