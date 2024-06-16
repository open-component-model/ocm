package logopts

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/config"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var _ = Describe("log configuration", func() {
	It("provides forward config", func() {
		ctx := clictx.New()

		opts := Options{
			ConfigFragment: ConfigFragment{
				LogLevel: "debug",
				LogKeys: []string{
					"tag=trace",
					"/realm=info",
					"/+all=info",
				},
			},
		}

		MustBeSuccessful(opts.Configure(ctx.OCMContext(), ctx.LoggingContext()))
		Expect(opts.LogForward).NotTo(BeNil())

		data := Must(yaml.Marshal(opts.LogForward))
		fmt.Printf("%s\n", string(data))
		logctx := logging.NewWithBase(nil)
		MustBeSuccessful(config.Configure(logctx, opts.LogForward))

		Expect(logctx.GetDefaultLevel()).To(Equal(logging.DebugLevel))
		Expect(logctx.Logger(logging.NewTag("tag")).Enabled(logging.TraceLevel)).To(BeTrue())
		Expect(logctx.Logger(logging.NewRealm("all")).Enabled(logging.DebugLevel)).To(BeFalse())
		Expect(logctx.Logger(logging.NewRealm("all/test")).Enabled(logging.DebugLevel)).To(BeFalse())
		Expect(logctx.Logger(logging.NewRealm("realm")).Enabled(logging.InfoLevel)).To(BeTrue())
		Expect(logctx.Logger(logging.NewRealm("realm")).Enabled(logging.DebugLevel)).To(BeFalse())
		Expect(logctx.Logger(logging.NewRealm("realm/test")).Enabled(logging.InfoLevel)).To(BeTrue())
		Expect(logctx.Logger(logging.NewRealm("realm/test")).Enabled(logging.DebugLevel)).To(BeTrue())
	})
})
