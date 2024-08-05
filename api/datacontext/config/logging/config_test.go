package logging_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/utils/logging/testhelper"

	"github.com/mandelsoft/logging"
	"github.com/tonglil/buflogr"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
	logcfg "ocm.software/ocm/api/datacontext/config/logging"
	log "ocm.software/ocm/api/utils/logging"
)

var _ = Describe("logging configuration", func() {
	var ctx datacontext.AttributesContext
	var cfg config.Context
	var buf bytes.Buffer
	var orig logging.Context

	BeforeEach(func() {
		orig = logging.DefaultContext().(*logging.ContextReference).Context
		logging.SetDefaultContext(logging.NewDefault())
		log.SetContext(nil)
		ctx = datacontext.New(nil)
		cfg = config.WithSharedAttributes(ctx).New()

		buf.Reset()
		def := buflogr.NewWithBuffer(&buf)
		ctx.LoggingContext().SetBaseLogger(def)
	})

	AfterEach(func() {
		// logging.SetDefaultContext(orig)
	})
	_ = cfg
	_ = orig

	It("just logs with defaults", func() {
		LogTest(ctx)

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
	})

	It("just logs with settings from default context", func() {
		logging.DefaultContext().AddRule(logging.NewConditionRule(logging.DebugLevel))
		LogTest(ctx)

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
	})

	It("just logs with settings from default context", func() {
		logging.DefaultContext().AddRule(logging.NewConditionRule(logging.DebugLevel))
		LogTest(cfg)

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
	})

	It("just logs with settings for root context", func() {
		spec := `
type: ` + logcfg.ConfigTypeV1 + `
contextType: ` + datacontext.CONTEXT_TYPE + `
settings:
  rules:
  - rule:
      level: Debug
`
		_, err := cfg.ApplyData([]byte(spec), nil, "testconfig")
		Expect(err).To(Succeed())
		LogTest(ctx)
		LogTest(cfg, "cfg")

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
V[4] cfgdebug realm ocm
V[3] cfginfo realm ocm
V[2] cfgwarn realm ocm
ERROR <nil> cfgerror realm ocm
`))
	})

	It("just logs with settings for root context by context provider", func() {
		spec := `
type: ` + logcfg.ConfigTypeV1 + `
settings:
  rules:
  - rule:
      level: Debug
`
		_, err := cfg.ApplyData([]byte(spec), nil, "testconfig")
		Expect(err).To(Succeed())

		LogTest(ctx)
		LogTest(cfg, "cfg")

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
V[4] cfgdebug realm ocm
V[3] cfginfo realm ocm
V[2] cfgwarn realm ocm
ERROR <nil> cfgerror realm ocm
`))
	})

	It("just logs with settings for config context", func() {
		spec := `
type: ` + logcfg.ConfigTypeV1 + `
contextType: ` + config.CONTEXT_TYPE + `
settings:
  rules:
  - rule:
      level: Debug
`
		_, err := cfg.ApplyData([]byte(spec), nil, "testconfig")
		Expect(err).To(Succeed())

		LogTest(ctx)
		LogTest(cfg, "cfg")

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
V[4] cfgdebug realm ocm
V[3] cfginfo realm ocm
V[2] cfgwarn realm ocm
ERROR <nil> cfgerror realm ocm
`))
	})

	Context("default logging", func() {
		spec1 := `
type: ` + logcfg.ConfigTypeV1 + `
contextType: default
settings:
  rules:
  - rule:
      level: Debug
`
		spec2 := `
type: ` + logcfg.ConfigTypeV1 + `
contextType: default
settings:
  rules:
  - rule:
      level: Info
`
		spec3 := `
type: ` + logcfg.ConfigTypeV1 + `
contextType: default
extraId: extra
settings:
  rules:
  - rule:
      level: Debug
`

		var ctx logging.Context

		BeforeEach(func() {
			log.SetContext(logging.NewDefault())
			buf.Reset()
			def := buflogr.NewWithBuffer(&buf)
			ctx = log.Context()
			ctx.SetBaseLogger(def)
		})

		It("just logs with config", func() {
			_, err := cfg.ApplyData([]byte(spec1), nil, "spec1")
			Expect(err).To(Succeed())
			LogTest(ctx)

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
		})

		It("applies config once", func() {
			_, err := cfg.ApplyData([]byte(spec1), nil, "spec1")
			Expect(err).To(Succeed())
			_, err = cfg.ApplyData([]byte(spec2), nil, "spec2")
			Expect(err).To(Succeed())
			_, err = cfg.ApplyData([]byte(spec1), nil, "spec1.2")
			Expect(err).To(Succeed())

			LogTest(ctx)

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
		})
		It("re-applies config with extra id", func() {
			_, err := cfg.ApplyData([]byte(spec1), nil, "spec1")
			Expect(err).To(Succeed())
			_, err = cfg.ApplyData([]byte(spec2), nil, "spec2")
			Expect(err).To(Succeed())
			_, err = cfg.ApplyData([]byte(spec3), nil, "spec3")
			Expect(err).To(Succeed())

			LogTest(ctx)

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
		})
	})
})
