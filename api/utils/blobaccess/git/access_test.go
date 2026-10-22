package git_test

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
		ctx     ocm.Context
		url     string
		repoDir string
		commit  plumbing.Hash
		parent  plumbing.Hash
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
			repoDir = GinkgoT().TempDir() + filepath.PathSeparatorString + "repo"

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
			Expect(wt.AddWithOptions(&git.AddOptions{All: true})).To(Succeed())
			parent, err = wt.Commit("OCM Test Parent", &git.CommitOptions{
				Author: &object.Signature{
					Name:  "OCM Test",
					Email: "dummy@ocm.software",
					When:  time.Now(),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(repoDir, "file_in_repo"), []byte("second revision\n"), 0o600)).To(Succeed())
			Expect(wt.AddWithOptions(&git.AddOptions{All: true})).To(Succeed())
			commit, err = wt.Commit("OCM Test Commit", &git.CommitOptions{
				Author: &object.Signature{Name: "OCM Test", Email: "dummy@ocm.software", When: time.Now()},
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
			expectArchiveFiles(files, "file_in_repo", "file_in_dir_in_repo")

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

			Expect(b.Digest()).To(Equal(b2.Digest()))
			bytes1 := Must(io.ReadAll(data1))
			bytes2 := Must(io.ReadAll(data2))
			Expect(data1.Close()).To(Succeed())
			Expect(data2.Close()).To(Succeed())
			Expect(bytes1).To(Equal(bytes2))

			cmd := exec.Command("git", "-C", repoDir, "gc", "--aggressive")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred(), string(output))
			b3 := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b3)
			Expect(b3.Digest()).To(Equal(b.Digest()))
		})

		It("contains a complete Git history with a detached HEAD", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			dir := GinkgoT().TempDir()
			out := Must(projectionfs.New(osfs.New(), dir))
			reader := Must(b.Reader())
			Expect(tarutils.UnzipTarToFs(out, reader)).To(Succeed())
			Expect(reader.Close()).To(Succeed())
			head := Must(os.ReadFile(filepath.Join(dir, ".git", "HEAD")))
			Expect(string(head)).To(Equal(commit.String() + "\n"))

			cmd := exec.Command("git", "-C", dir, "fsck", "--strict")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred(), string(output))
			cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD^", "HEAD")
			output, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred(), string(output))
			Expect(strings.Fields(string(output))).To(Equal([]string{parent.String(), commit.String()}))
		})

		It("creates an archive without host-specific tar headers", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			reader := Must(b.Reader())
			defer Close(reader)
			gz := Must(gzip.NewReader(reader))
			tr := tar.NewReader(gz)
			for {
				h, err := tr.Next()
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(h.Uid).To(BeZero(), h.Name)
				Expect(h.Gid).To(BeZero(), h.Name)
				Expect(h.Uname).To(BeEmpty(), h.Name)
				Expect(h.Gname).To(BeEmpty(), h.Name)
				Expect(h.Mode).To(BeElementOf(int64(0o644), int64(0o755)), h.Name)
			}
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
			expectArchiveFiles(files, "README")
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
			expectArchiveFiles(files, "README")
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
			expectArchiveFiles(files, "README", "CONTRIBUTING.md")
		})
	})

})

func expectArchiveFiles(files []string, worktree ...string) {
	var history, contents []string
	for _, name := range files {
		if name == ".git" || strings.HasPrefix(name, ".git/") {
			history = append(history, name)
		} else {
			contents = append(contents, name)
		}
	}
	Expect(contents).To(ConsistOf(worktree))
	Expect(history).To(ContainElements(".git", ".git/HEAD", ".git/objects", ".git/objects/pack", ".git/refs"))
	Expect(history).To(HaveLen(7))
	packPattern := regexp.MustCompile(`^\.git/objects/pack/pack-[0-9a-f]{40}\.pack$`)
	indexPattern := regexp.MustCompile(`^\.git/objects/pack/pack-[0-9a-f]{40}\.idx$`)
	Expect(history[4]).To(MatchRegexp(indexPattern.String()))
	Expect(history[5]).To(MatchRegexp(packPattern.String()))
	Expect(strings.TrimSuffix(history[4], ".idx")).To(Equal(strings.TrimSuffix(history[5], ".pack")))
}
