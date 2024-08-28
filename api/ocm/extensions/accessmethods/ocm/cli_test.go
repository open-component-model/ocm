package ocm_test

import (
	"github.com/mandelsoft/goutils/sliceutils"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/goutils/transformer"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ocm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var _ = Describe("OCM access CLI Test Environment", func() {
	ctx := cpi.DefaultContext()

	Context("cli options", func() {
		It("handles access options", func() {
			at := ctx.AccessMethods().GetType(ocm.TypeV1)
			Expect(at).NotTo(BeNil())

			h := at.ConfigOptionTypeSetHandler()
			Expect(h).NotTo(BeNil())
			Expect(h.GetName()).To(Equal(ocm.Type))

			ot := h.OptionTypes()
			Expect(len(ot)).To(Equal(4))

			opts := h.CreateOptions()
			Expect(sliceutils.Transform(opts.Options(), transformer.GetName[flagsets.Option, string])).To(ConsistOf(
				"accessComponent", "accessVersion", "accessRepository", "accessIdentityPath"))

			fs := &pflag.FlagSet{}
			fs.SortFlags = true
			opts.AddFlags(fs)

			Expect("\n" + fs.FlagUsages()).To(Equal(`
      --accessComponent string                component for access specification
      --accessIdentityPath {<name>=<value>}   identity path for access specification
      --accessRepository string               repository URL
      --accessVersion string                  version for access specification
`))

			MustBeSuccessful(fs.Parse([]string{
				"--accessRepository", "ghcr.io/open-component-model/ocm",
				"--accessComponent", COMP1,
				"--accessVersion", VERS,
				"--accessIdentityPath", "name=rsc1",
				"--accessIdentityPath", "other=value",
				"--accessIdentityPath", "name=rsc2",
			}))

			cfg := flagsets.Config{}
			MustBeSuccessful(h.ApplyConfig(opts, cfg))
			Expect(cfg).To(YAMLEqual(`
  component: acme.org/test1
  version: v1
  ocmRepository:
    type: OCIRegistry
    baseUrl: ghcr.io
    componentNameMapping: urlPath
    subPath: open-component-model/ocm
  resourceRef:
  - name: rsc1
    other: value
  - name: rsc2
`))
		})
	})
})
