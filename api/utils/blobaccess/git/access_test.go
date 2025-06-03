package git_test

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"

	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	gitblob "ocm.software/ocm/api/utils/blobaccess/git"
	"ocm.software/ocm/api/utils/tarutils"
)

//go:embed testdata/repo
var testData embed.FS

var _ = Describe("git Blob Access", func() {
	var (
		ctx ocm.Context
		url string
	)

	ctx = ocm.New()

	BeforeEach(func() {
		tempVFS, err := projectionfs.New(osfs.New(), GinkgoT().TempDir())
		Expect(err).ToNot(HaveOccurred())
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: ".", Filesystem: tempVFS})
		vfsattr.Set(ctx, tempVFS)
	})

	Context("git filesystem repository", func() {
		BeforeEach(func() {
			repoDir := GinkgoT().TempDir() + filepath.PathSeparatorString + "repo"

			repo, err := git.PlainInit(repoDir, false)
			Expect(err).ToNot(HaveOccurred())

			repoBase := filepath.Join("testdata", "repo")
			repoTestData, err := testData.ReadDir(repoBase)
			Expect(err).ToNot(HaveOccurred())

			var process func(base string, entries []fs.DirEntry)
			process = func(base string, infos []fs.DirEntry) {
				for _, entry := range infos {
					path := filepath.Join(base, entry.Name())
					repoPath := filepath.Join(repoDir, entry.Name())

					file, err := testData.Open(path)
					Expect(err).ToNot(HaveOccurred())
					fi, err := file.Stat()
					Expect(err).ToNot(HaveOccurred())
					if fi.IsDir() {
						Expect(os.MkdirAll(repoPath, 0o700)).ToNot(HaveOccurred())
						entries, err := testData.ReadDir(path)
						Expect(err).ToNot(HaveOccurred())
						process(filepath.Join(base, fi.Name()), entries)
						continue
					}

					fileInRepo, err := os.OpenFile(
						repoPath,
						os.O_CREATE|os.O_RDWR|os.O_TRUNC,
						0o600,
					)
					Expect(err).ToNot(HaveOccurred())

					_, err = io.Copy(fileInRepo, file)
					Expect(err).ToNot(HaveOccurred())

					Expect(fileInRepo.Close()).To(Succeed())
					Expect(file.Close()).To(Succeed())
				}
			}
			process(repoBase, repoTestData)

			wt, err := repo.Worktree()
			Expect(err).ToNot(HaveOccurred())
			Expect(wt.AddGlob("*")).To(Succeed())
			_, err = wt.Commit("OCM Test Commit", &git.CommitOptions{
				Author: &object.Signature{
					Name:  "OCM Test",
					Email: "dummy@ocm.software",
					When:  time.Now(),
				},
			})
			Expect(err).ToNot(HaveOccurred())

			url = fmt.Sprintf("file://%s", repoDir)
		})

		It("blobaccess for simple repository", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("file_in_repo", "file_in_dir_in_repo"))

			data1, err := b.Reader()
			Expect(err).ToNot(HaveOccurred())

			By("testing equivalence of byte streams from equivalent accesses")
			b2 := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b2)
			data2, err := b2.Reader()
			Expect(err).ToNot(HaveOccurred())

			Expect(data1).ToNot(BeIdenticalTo(data2))
		})
	})

	Context("git http repository", func() {
		host := "github.com:443"
		reachable := PingTCPServer(host, time.Second) == nil
		BeforeEach(func() {
			if !reachable {
				Skip(fmt.Sprintf("no connection to %s, skipping test connection to remote", url))
			}
			// This repo is a public repo owned by the Github Kraken Bot, so its as good of a public available
			// example as any.
			url = fmt.Sprintf("https://%s/octocat/Hello-World.git", host)
		})

		It("hello world from github without explicit branch", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("README"))
		})

		It("hello world from github with master branch", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
				gitblob.WithRef(plumbing.Master.String()),
			))
			defer Close(b)
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("README"))
		})

		It("hello world from github with custom ref", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
				gitblob.WithRef("refs/heads/test"),
			))
			defer Close(b)
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("README", "CONTRIBUTING.md"))
		})
	})

})
