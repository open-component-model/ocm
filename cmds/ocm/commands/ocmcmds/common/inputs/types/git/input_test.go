package git_test

import (
	"fmt"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
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
