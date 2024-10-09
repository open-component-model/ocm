package git_test

import (
	"io"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/tarutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH    = "test.ca"
	VERSION = "v1"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())

		Expect(env.Execute(
			"create",
			"ca",
			"-ft",
			"directory",
			"test.de/x",
			VERSION,
			"--provider",
			"ocm",
			"--file",
			ARCH,
			"--scheme",
			"ocm.software/v3alpha1",
		)).To(Succeed())
	})

	AfterEach(func() {
		Expect(env.Cleanup()).To(Succeed())
	})

	It("add git repo described by access type specification", func() {
		meta := `
name: hello-world
type: git
`
		Expect(env.Execute(
			"add", "resources",
			"--file", ARCH,
			"--resource", meta,
			"--accessType", "git",
			"--accessRepository", "https://github.com/octocat/Hello-World.git",
			"--reference", "refs/heads/master",
			"--commit", "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
			"--version", "0.0.1",
		)).To(Succeed())

		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))
	})

	It("add git repo described by cli options through blob access via input described in file", func() {
		meta := `
name: hello-world
type: git
`
		Expect(env.Execute(
			"add", "resources",
			"--file", ARCH,
			"--resource", meta,
			"--inputType", "git",
			"--inputVersion", "refs/heads/master",
			"--inputRepository", "https://github.com/octocat/Hello-World.git",
		)).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
		Expect(access.MediaType).To(Equal(mime.MIME_TGZ))
		fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
		Expect(fi.Size()).To(Equal(int64(106)))

		Expect(tarutils.ExtractArchiveToFs(env.FileSystem(), env.Join(ARCH, "blobs", access.LocalReference), env.FileSystem())).To(Succeed())

		readMeFi := Must(env.FileSystem().Stat("README"))
		Expect(readMeFi.Size()).To(Equal(int64(13)))
		readMe := Must(env.FileSystem().OpenFile("README", os.O_RDONLY, 0o400))
		defer readMe.Close()
		Expect(string(Must(io.ReadAll(readMe)))).To(Equal("Hello World!\n"))
	})

	It("add git repo described by cli options through blob access via input", func() {
		meta := `
name: hello-world
type: git
`
		Expect(env.Execute(
			"add", "resources",
			"--file", ARCH,
			"--resource", meta,
			"--inputType", "git",
			"--inputVersion", "refs/heads/master",
			"--inputRepository", "https://github.com/octocat/Hello-World.git",
		)).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
		Expect(access.MediaType).To(Equal(mime.MIME_TGZ))
		fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
		Expect(fi.Size()).To(Equal(int64(106)))

		Expect(tarutils.ExtractArchiveToFs(env.FileSystem(), env.Join(ARCH, "blobs", access.LocalReference), env.FileSystem())).To(Succeed())

		readMeFi := Must(env.FileSystem().Stat("README"))
		Expect(readMeFi.Size()).To(Equal(int64(13)))
		readMe := Must(env.FileSystem().OpenFile("README", os.O_RDONLY, 0o400))
		defer readMe.Close()
		Expect(string(Must(io.ReadAll(readMe)))).To(Equal("Hello World!\n"))
	})
})
