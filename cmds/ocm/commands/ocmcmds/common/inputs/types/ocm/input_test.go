package ocm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	. "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ocm"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	COMP = "acme.org/test"
	VERS = "v1"
	ARCH = "/ca"

	REPO    = "/ctf"
	REFCOMP = "acme.org/ref"
	REFRSC  = "res"
)

var _ = Describe("Input Type", func() {
	Context("spec", func() {
		var env *InputTest

		BeforeEach(func() {
			env = NewInputTest(ocm.TYPE)
		})

		It("simple spec", func() {
			env.Set(options.RepositoryOption, "ghcr.io/open-component-model")
			env.Set(options.ComponentOption, COMP)
			env.Set(options.VersionOption, VERS)
			env.Set(options.IdentityPathOption, "name=test")
			env.Check(&ocm.Spec{
				InputSpecBase: inputs.InputSpecBase{},
				OCMRepository: Must(cpi.ToGenericRepositorySpec(ocireg.NewRepositorySpec("ghcr.io/open-component-model"))),
				Component:     COMP,
				Version:       VERS,
				ResourceRef:   v1.NewResourceRef(v1.NewIdentity("test")),
			})
		})
	})

	Context("compose", func() {
		var env *TestEnv

		BeforeEach(func() {
			env = NewTestEnv(TestData())

			env.OCMCommonTransport(REPO, accessio.FormatDirectory, func() {
				env.ComponentVersion(REFCOMP, VERS, func() {
					env.Resource(REFRSC, VERS, resourcetypes.PLAIN_TEXT, v1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "test data")
					})
				})
			})

			Expect(env.Execute("create", "ca", "-ft", "directory", COMP, VERS, "--provider", "acme.org", "--file", ARCH)).To(Succeed())
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("adds ocm resource", func() {
			Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/resources1.yaml")).To(Succeed())

			ca := Must(comparch.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
			defer Close(ca)

			r := Must(ca.GetResourceByIndex(0))
			Expect(Must(r.Access()).GetKind()).To(Equal(localblob.Type))

			data := Must(ocmutils.GetResourceData(r))
			Expect(string(data)).To(Equal("test data"))
		})
	})
})
