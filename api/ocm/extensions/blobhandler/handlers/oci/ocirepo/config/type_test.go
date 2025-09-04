package config_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cfgctx "ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo/config"
)

var _ = Describe("Test Environment", func() {
	var opts *config.UploadOptions

	BeforeEach(func() {
		opts = &config.UploadOptions{
			PreferRelativeAccess: true,
			Repositories: []string{
				"localhost",
				"localhost:5000",
				"host",
				"other:5000",
			},
		}
	})

	Context("options", func() {
		It("check without repos", func() {
			Expect((&config.UploadOptions{PreferRelativeAccess: true}).PreferRelativeAccessFor("host")).To(BeTrue())
			Expect((&config.UploadOptions{PreferRelativeAccess: false}).PreferRelativeAccessFor("host")).To(BeFalse())
		})

		It("check for host", func() {
			Expect(opts.PreferRelativeAccessFor("localhost")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("localhost:5000")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("localhost:6000")).To(BeTrue())

			Expect(opts.PreferRelativeAccessFor("host")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("host:5000")).To(BeTrue())

			Expect(opts.PreferRelativeAccessFor("other:5000")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("other")).To(BeFalse())

			Expect(opts.PreferRelativeAccessFor("any")).To(BeFalse())
		})
	})

	Context("config object", func() {
		var cfg cfgctx.Context

		BeforeEach(func() {
			cfg = cfgctx.New(datacontext.MODE_DEFAULTED)
		})

		It("configures", func() {
			o := config.New()
			o.UploadOptions = *opts

			MustBeSuccessful(cfg.ApplyConfig(o, "manual"))

			Expect(opts.PreferRelativeAccessFor("localhost")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("localhost:5000")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("localhost:6000")).To(BeTrue())

			Expect(opts.PreferRelativeAccessFor("host")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("host:5000")).To(BeTrue())

			Expect(opts.PreferRelativeAccessFor("other:5000")).To(BeTrue())
			Expect(opts.PreferRelativeAccessFor("other")).To(BeFalse())

			Expect(opts.PreferRelativeAccessFor("any")).To(BeFalse())
		})
	})
})
