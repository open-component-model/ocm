package git_test

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "embed"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/opencontainers/go-digest"

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

			Expect(b.Digest()).To(Equal(b2.Digest()))
			bytes1 := Must(io.ReadAll(data1))
			bytes2 := Must(io.ReadAll(data2))
			Expect(data1.Close()).To(Succeed())
			Expect(data2.Close()).To(Succeed())
			Expect(bytes1).To(Equal(bytes2))
		})

		It("contains a complete Git history with a detached HEAD", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithFullHistory(),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			expectArchiveFiles(Must(tarutils.ListArchiveContentFromReader(Must(b.Reader()))), "file_in_repo", "file_in_dir_in_repo")
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

			cmd = exec.Command("git", "-C", repoDir, "gc", "--aggressive")
			output, err = cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred(), string(output))
			repacked := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithFullHistory(),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(repacked)
			Expect(repacked.Digest()).To(Equal(b.Digest()))
		})

		It("creates an archive without host-specific tar headers", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(url),
				gitblob.WithFullHistory(),
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

	Context("archive layout of OCM v2", func() {
		// The fixture has fixed commit metadata and is built without ambient Git
		// configuration, so it always has the same commit and hence the same archive.
		const (
			fixtureCommit = "0c4a69dc63c3a99b921e85c2bfc1023dd0e507b9"
			fixtureParent = "e9449c3d074078c398f0a313a1d6b1173e789de9"
			// archiveDigest is the SHA-256 of the uncompressed archive OCM v2 builds
			// for fixtureCommit, which this implementation must reproduce byte for
			// byte: a resource created here is verified against this digest by v2,
			// and one created by v2 against the digest produced here. Update it only
			// from an archive v2 produced for the fixture, never from the bytes this
			// implementation happens to produce.
			archiveDigest = "sha256:692de62dbd8a107bf0c1f239deae4361c0f3c6183fce00de04773cf0bd7b2deb"
		)

		var layoutURL string

		BeforeEach(func() {
			dir := filepath.Join(GinkgoT().TempDir(), "fixture")
			Expect(os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "sub", "deep", "file.txt"), []byte("nested\n"), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh\necho hi\n"), 0o700)).To(Succeed())
			Expect(os.Symlink("README.md", filepath.Join(dir, "link.md"))).To(Succeed())

			const first = "2020-01-01T00:00:00+0000"
			fixtureGit(dir, first, "init", "-q", "-b", "main", ".")
			fixtureGit(dir, first, "add", "-A")
			fixtureGit(dir, first, "-c", "commit.gpgsign=false", "commit", "-q", "-m", "first")

			const second = "2020-01-02T00:00:00+0000"
			Expect(os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\nsecond\n"), 0o600)).To(Succeed())
			fixtureGit(dir, second, "add", "-A")
			fixtureGit(dir, second, "-c", "commit.gpgsign=false", "commit", "-q", "-m", "second")

			layoutURL = fmt.Sprintf("file://%s", dir)
		})

		It("digests the bytes OCM v2 digests for the same commit", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(layoutURL),
				gitblob.WithCommit(fixtureCommit),
				gitblob.WithFullHistory(),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			reader := Must(b.Reader())
			defer Close(reader)
			unzipped := Must(gzip.NewReader(reader))

			// The compressed digest additionally depends on the gzip implementation,
			// which neither side guarantees; the archive itself must match.
			Expect(digest.SHA256.FromReader(unzipped)).To(Equal(digest.Digest(archiveDigest)))
		})

		It("writes the Git storage of an OCM v2 archive", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(layoutURL),
				gitblob.WithCommit(fixtureCommit),
				gitblob.WithFullHistory(),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			expectArchiveFiles(Must(tarutils.ListArchiveContentFromReader(Must(b.Reader()))),
				"README.md", "link.md", "run.sh", "sub", "sub/deep", "sub/deep/file.txt")

			dir := GinkgoT().TempDir()
			out := Must(projectionfs.New(osfs.New(), dir))
			reader := Must(b.Reader())
			Expect(tarutils.UnzipTarToFs(out, reader)).To(Succeed())
			Expect(reader.Close()).To(Succeed())

			Expect(string(Must(os.ReadFile(filepath.Join(dir, ".git", "HEAD"))))).To(Equal(fixtureCommit + "\n"))
			Expect(string(Must(os.ReadFile(filepath.Join(dir, ".git", "config"))))).To(
				Equal("[core]\n\tbare = false\n\trepositoryformatversion = 1\n\tfilemode = true\n"))

			output := Must(exec.Command("git", "-C", dir, "rev-parse", "HEAD^", "HEAD").CombinedOutput())
			Expect(strings.Fields(string(output))).To(Equal([]string{fixtureParent, fixtureCommit}))
		})

		It("writes the packfile index and reverse index git writes", func() {
			b := Must(gitblob.BlobAccess(
				gitblob.WithURL(layoutURL),
				gitblob.WithCommit(fixtureCommit),
				gitblob.WithFullHistory(),
				gitblob.WithLoggingContext(ctx),
				gitblob.WithCachingContext(ctx),
			))
			defer Close(b)
			dir := GinkgoT().TempDir()
			out := Must(projectionfs.New(osfs.New(), dir))
			reader := Must(b.Reader())
			Expect(tarutils.UnzipTarToFs(out, reader)).To(Succeed())
			Expect(reader.Close()).To(Succeed())

			packs := Must(filepath.Glob(filepath.Join(dir, ".git", "objects", "pack", "pack-*.pack")))
			Expect(packs).To(HaveLen(1))

			// go-git writes the index itself and cannot write a reverse index at all,
			// so both are compared against the files git derives from the packfile.
			check := GinkgoT().TempDir()
			fixtureGit(check, "", "init", "-q", ".")
			Expect(os.WriteFile(filepath.Join(check, "candidate.pack"), Must(os.ReadFile(packs[0])), 0o600)).To(Succeed())
			fixtureGit(check, "", "index-pack", "--rev-index", "-o", "candidate.idx", "candidate.pack")

			for _, extension := range []string{".idx", ".rev"} {
				Expect(Must(os.ReadFile(strings.TrimSuffix(packs[0], ".pack")+extension))).To(
					Equal(Must(os.ReadFile(filepath.Join(check, "candidate"+extension)))), extension)
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

// expectArchiveFiles asserts that the archive consists of the Git storage OCM v2
// writes, in the order v2 writes it, plus the given worktree entries. The order is
// part of the contract: both implementations must produce the same archive bytes.
func expectArchiveFiles(files []string, worktree ...string) {
	var storage, contents []string
	for _, name := range files {
		if name == ".git" || strings.HasPrefix(name, ".git/") {
			storage = append(storage, name)
		} else {
			contents = append(contents, name)
		}
	}
	Expect(contents).To(ConsistOf(worktree))
	Expect(storage).To(HaveLen(12))
	pack := strings.TrimSuffix(storage[6], ".idx")
	Expect(pack).To(MatchRegexp(`^\.git/objects/pack/pack-[0-9a-f]{40}$`))
	Expect(storage).To(Equal([]string{
		".git", ".git/HEAD", ".git/config",
		".git/objects", ".git/objects/info", ".git/objects/pack",
		pack + ".idx", pack + ".pack", pack + ".rev",
		".git/refs", ".git/refs/heads", ".git/refs/tags",
	}))
}

// fixtureGit runs git in dir with a fixed identity and commit date and without
// any ambient configuration, so the fixture has the same objects on every machine.
func fixtureGit(dir, date string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + dir,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=OCM Test",
		"GIT_AUTHOR_EMAIL=dummy@ocm.software",
		"GIT_COMMITTER_NAME=OCM Test",
		"GIT_COMMITTER_EMAIL=dummy@ocm.software",
		"GIT_AUTHOR_DATE=" + date,
		"GIT_COMMITTER_DATE=" + date,
	}
	output, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), string(output))
}
