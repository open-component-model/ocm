package git_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/tarutils"
)

const (
	ARCH        = "test.ctf"
	CONSTRUCTOR = "component-constructor.yaml"
	VERSION     = "v1"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		Expect(env.Cleanup()).To(Succeed())
	})

	It("packages the git history for every input type name", func() {
		repository := fixtureRepository()

		digests := map[string]string{}
		for _, typ := range []string{"git", "git/v1", "Git", "Git/v1"} {
			constructor := fmt.Sprintf(`---
name: test.de/x
version: %s
provider:
  name: ocm
resources:
- name: repo
  type: directoryTree
  version: 0.0.1
  input:
    type: %s
    ref: refs/heads/main
    repository: %s
`, VERSION, typ, repository)
			Expect(env.WriteFile(CONSTRUCTOR, []byte(constructor), os.ModePerm)).To(Succeed())
			Expect(env.Execute("add", "cv", "--file", ARCH, "--create", "--force", "--type", "directory", CONSTRUCTOR)).To(Succeed())

			ctx := ocm.New()
			vfsattr.Set(ctx, env.FileSystem())
			repo := Must(ctf.Open(ctx, ctf.ACC_READONLY, ARCH, 0o400, accessio.FormatDirectory, accessio.PathFileSystem(env.FileSystem())))
			c := Must(repo.LookupComponent("test.de/x"))
			cv := Must(c.LookupVersion(VERSION))

			r := Must(cv.GetResource(metav1.Identity{"name": "repo"}))
			m := Must(r.AccessMethod())
			archive := Must(m.Get())
			files := Must(tarutils.ListArchiveContentFromReader(bytes.NewReader(archive)))
			Expect(files).To(ContainElements(".git/HEAD", "file"), typ)
			Expect(files).To(ContainElement(MatchRegexp(`^\.git/objects/pack/pack-[0-9a-f]{40}\.pack$`)), typ)
			digests[typ] = digest.FromBytes(archive).String()

			Expect(errors.Join(m.Close(), cv.Close(), c.Close(), repo.Close())).To(Succeed())
		}

		// The type names are alternative spellings of the same input, in OCM v1 as
		// well as in OCM v2, so they must all package the same archive.
		Expect(digests).To(HaveLen(4))
		for typ, d := range digests {
			Expect(d).To(Equal(digests["Git/v1"]), typ)
		}
	})

	It("add git repo described by access type specification", func() {
		constructor := fmt.Sprintf(`---
name: test.de/x
version: %s
provider:
  name: ocm
resources:
- name: hello-world
  type: git
  version: 0.0.1
  access:
    type: git
    commit: "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d"
    ref: refs/heads/master
    repository: https://github.com/octocat/Hello-World.git
`, VERSION)
		Expect(
			env.WriteFile(CONSTRUCTOR, []byte(constructor), os.ModePerm),
		).To(Succeed())

		Expect(env.Execute(
			"add",
			"cv",
			"--create",
			"--file",
			ARCH,
			"--force",
			"--type",
			"directory",
			CONSTRUCTOR,
		)).To(Succeed())

		ctx := ocm.New()
		vfsattr.Set(ctx, env.FileSystem())
		r := Must(ctf.Open(ctx, ctf.ACC_READONLY, ARCH, 0o400, accessio.FormatDirectory, accessio.PathFileSystem(env.FileSystem())))
		DeferCleanup(r.Close)

		c := Must(r.LookupComponent("test.de/x"))
		DeferCleanup(c.Close)
		cv := Must(c.LookupVersion(VERSION))
		DeferCleanup(cv.Close)
		cd := cv.GetDescriptor()
		Expect(len(cd.Resources)).To(Equal(1))
	})

	It("add git repo described by cli options through blob access via input described in file", func() {
		constructor := fmt.Sprintf(`---
name: test.de/x
version: %s
provider:
  name: ocm
resources:
- name: hello-world
  type: git
  version: 0.0.1
  input:
    type: git
    ref: refs/heads/master
    commit: "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d"
    repository: https://github.com/octocat/Hello-World.git
`, VERSION)
		Expect(
			env.WriteFile(CONSTRUCTOR, []byte(constructor), os.ModePerm),
		).To(Succeed())

		Expect(env.Execute(
			"add",
			"cv",
			"--file",
			ARCH,
			"--create",
			"--force",
			"--type",
			"directory",
			CONSTRUCTOR,
		)).To(Succeed())

		ctx := ocm.New()
		vfsattr.Set(ctx, env.FileSystem())
		r := Must(ctf.Open(ctx, ctf.ACC_READONLY, ARCH, 0o400, accessio.FormatDirectory, accessio.PathFileSystem(env.FileSystem())))
		DeferCleanup(r.Close)

		c := Must(r.LookupComponent("test.de/x"))
		DeferCleanup(c.Close)
		cv := Must(c.LookupVersion(VERSION))
		DeferCleanup(cv.Close)
		cd := cv.GetDescriptor()
		Expect(len(cd.Resources)).To(Equal(1))
	})
})

// fixtureRepository creates a local Git repository with one commit and returns its
// URL. The repository lives on the host filesystem, as git transports read it.
func fixtureRepository() string {
	dir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(dir, "file"), []byte("content\n"), 0o600)).To(Succeed())

	for _, args := range [][]string{
		{"init", "-q", "-b", "main", "."},
		{"add", "-A"},
		{"-c", "user.name=OCM Test", "-c", "user.email=dummy@ocm.software", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "fixture"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), string(output))
	}
	return "file://" + dir
}
