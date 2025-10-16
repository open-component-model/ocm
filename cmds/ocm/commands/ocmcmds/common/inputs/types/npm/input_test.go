package npm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/tech/npm/npmtest"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	. "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/npm"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH    = "test.ca"
	VERSION = "v1"
)

var _ = Describe("Input Type", func() {
	Context("spec", func() {
		var env *InputTest

		BeforeEach(func() {
			env = NewInputTest(npm.TYPE)
		})

		It("simple fetch", func() {
			env.Set(options.RepositoryOption, "https://registry.npmjs.org")
			env.Set(options.PackageOption, "yargs")
			env.Set(options.VersionOption, "17.7.1")
			env.Check(&npm.Spec{
				InputSpecBase: inputs.InputSpecBase{},
				Registry:      "https://registry.npmjs.org",
				Package:       "yargs",
				Version:       "17.7.1",
			})
		})
	})

	Context("remote", func() {
		var env *TestEnv

		BeforeEach(func() {
			env = NewTestEnv(npmtest.TestData())
			Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", ARCH)).To(Succeed())
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("add npm from npm registry described by cli options", func() {
			meta := `
name: testdata
type: npmPackage
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "npm",
				"--inputRepository", "https://registry.npmjs.org", "--package", "discord.js",
				"--inputVersion", "14.8.0")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))
			access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
			Expect(access.MediaType).To(Equal(mime.MIME_TGZ))
			fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
			Expect(fi.Size()).To(Equal(int64(313366)))
		})

		It("add npm from file registry described by cli options", func() {
			meta := `
name: testdata
type: npmPackage
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "npm",
				"--inputRepository", "file://"+npmtest.NPMPATH, "--package", npmtest.PACKAGE,
				"--inputVersion", npmtest.VERSION)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))
			access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
			Expect(access.MediaType).To(Equal(mime.MIME_TGZ))
			fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
			Expect(fi.Size()).To(Equal(int64(npmtest.ARTIFACT_SIZE)))
		})
	})
})
