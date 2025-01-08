package check_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/ocmutils/check"
	"ocm.software/ocm/api/utils/accessio"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	ARCH    = "/tmp/ca"
	VERSION = "v1"
	COMP    = "test.de/x"
	COMP2   = "test.de/y"
	COMP3   = "test.de/z"
	COMP4   = "test.de/a"
)

var _ = Describe("Test Environment", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get checks references", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION)
		})

		spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
		repo := Must(env.OCMContext().RepositoryForSpec(spec))
		defer Close(repo, "repo")
		result := Must(check.Check().ForId(repo, common.NewNameVersion(COMP, VERSION)))
		Expect(result).To(BeNil())
	})

	Context("finds missing", func() {
		var repo ocm.Repository

		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERSION, func() {
					env.Reference("ref", COMP3, VERSION)
				})
				env.ComponentVersion(COMP2, VERSION, func() {
					env.Reference("ref", COMP3, VERSION)
				})
			})

			spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
			repo = Must(env.OCMContext().RepositoryForSpec(spec))
		})

		AfterEach(func() {
			MustBeSuccessful(repo.Close())
		})

		It("outputs table", func() {
			result := Must(check.Check().ForId(repo, common.NewNameVersion(COMP, VERSION)))

			Expect(result).NotTo(BeNil())
			Expect(json.Marshal(result)).To(YAMLEqual(`
missing:
 test.de/z:v1:
  - test.de/x:v1
  - test.de/z:v1
`))
		})
	})

	It("handles diamond", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref1", COMP2, VERSION)
				env.Reference("ref2", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP4, VERSION, func() {
			})
		})

		spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
		repo := Must(env.OCMContext().RepositoryForSpec(spec))
		defer Close(repo, "repo")
		result := Must(check.Check().ForId(repo, common.NewNameVersion(COMP, VERSION)))
		Expect(result).To(BeNil())
	})

	It("finds cycle", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION, func() {
				env.Reference("ref", COMP2, VERSION)
			})
		})

		spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
		repo := Must(env.OCMContext().RepositoryForSpec(spec))
		defer Close(repo, "repo")
		ExpectError(check.Check().ForId(repo, common.NewNameVersion(COMP, VERSION))).To(
			MatchError("component version recursion: use of test.de/z:v1 for test.de/x:v1->test.de/z:v1->test.de/y:v1"))
	})

	Context("finds non-local resources", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERSION, func() {
					env.Resource("rsc1", VERSION, resourcetypes.BLUEPRINT, v1.LocalRelation, func() {
						env.ModificationOptions(ocm.SkipDigest())
						env.Access(ociartifact.New("bla"))
					})
				})
			})
		})

		It("finds resources", func() {
			spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
			repo := Must(env.OCMContext().RepositoryForSpec(spec))
			defer Close(repo, "repo")
			result := Must(check.Check(check.LocalResourcesOnly()).ForId(repo, common.NewNameVersion(COMP, VERSION)))
			Expect(result).NotTo(BeNil())
			Expect(result).To(YAMLEqual(`
resources:
  - name: rsc1
`))
		})

		It("does not find resources", func() {
			spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
			repo := Must(env.OCMContext().RepositoryForSpec(spec))
			defer Close(repo, "repo")
			result := Must(check.Check().ForId(repo, common.NewNameVersion(COMP, VERSION)))
			Expect(result).To(BeNil())
		})
	})

	Context("finds non-local resources", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERSION, func() {
					env.Source("rsc1", VERSION, resourcetypes.BLUEPRINT, func() {
						env.ModificationOptions(ocm.SkipDigest())
						env.Access(ociartifact.New("bla"))
					})
				})
			})
		})

		It("finds sources", func() {
			spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
			repo := Must(env.OCMContext().RepositoryForSpec(spec))
			defer Close(repo, "repo")
			result := Must(check.Check(check.LocalSourcesOnly()).ForId(repo, common.NewNameVersion(COMP, VERSION)))
			Expect(result).NotTo(BeNil())
			Expect(result).To(YAMLEqual(`
sources:
  - name: rsc1
`))
		})

		It("does not find sources", func() {
			spec := Must(ctf.NewRepositorySpec(ctf.ACC_READONLY, ARCH, env))
			repo := Must(env.OCMContext().RepositoryForSpec(spec))
			defer Close(repo, "repo")
			result := Must(check.Check().ForId(repo, common.NewNameVersion(COMP, VERSION)))
			Expect(result).To(BeNil())
		})
	})
})
