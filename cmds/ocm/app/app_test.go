package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/tonglil/buflogr"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/extensions/attrs/mapocirepoattr"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/logging/testhelper"
)

var realm = logging.NewRealm("test")

func addTestCommands(ctx clictx.Context, cmd *cobra.Command) {
	if cmd != nil {
		c := &cobra.Command{
			Use:   "logtest",
			Short: "test log output",
			Run: func(cmd *cobra.Command, args []string) {
				testhelper.LoggerTest(ocmlog.Context().Logger(realm))
				testhelper.LoggerTest(ctx.LoggingContext().Logger(realm), "ctx")
			},
		}
		cmd.AddCommand(c)
	}
}

var _ = Describe("Test Environment", func() {
	var log bytes.Buffer
	var env *TestEnv
	var oldlog *ocmlog.StaticContext

	BeforeEach(func() {
		oldlog = ocmlog.Context()
		log.Reset()
		def := buflogr.NewWithBuffer(&log)
		n := logging.New(def)
		ocmlog.SetContext(n)
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
		ocmlog.SetContext(oldlog)
	})

	It("version gets the version", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("version")).To(Succeed())
		Expect(strings.HasPrefix(buf.String(), "{\"Major\":")).To(BeTrue())
		var m map[string]interface{}
		Expect(json.Unmarshal(buf.Bytes(), &m)).To(Succeed())
	})

	It("do logging", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-X", "plugindir=xxx", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[2] warn realm ocm realm test
ERROR <nil> error realm ocm realm test
V[2] ctxwarn realm ocm realm test
ERROR <nil> ctxerror realm ocm realm test
`))
	})

	It("sets logging", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-X", "plugindir=xxx", "-l", "Debug", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm realm test
V[3] info realm ocm realm test
V[2] warn realm ocm realm test
ERROR <nil> error realm ocm realm test
V[4] ctxdebug realm ocm realm test
V[3] ctxinfo realm ocm realm test
V[2] ctxwarn realm ocm realm test
ERROR <nil> ctxerror realm ocm realm test
`))
		ocmlog.Context().SetDefaultLevel(logging.WarnLevel)
	})

	It("sets logging by config", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "--logconfig", "@testdata/logcfg.yaml", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm realm test
V[3] info realm ocm realm test
V[2] warn realm ocm realm test
ERROR <nil> error realm ocm realm test
V[4] ctxdebug realm ocm realm test
V[3] ctxinfo realm ocm realm test
V[2] ctxwarn realm ocm realm test
ERROR <nil> ctxerror realm ocm realm test
`))
	})

	It("sets logging by config", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "--logconfig", `
defaultLevel: Warn
rules:
  - rule:
      level: Debug
      conditions:
        - realm: test`, "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm realm test
V[3] info realm ocm realm test
V[2] warn realm ocm realm test
ERROR <nil> error realm ocm realm test
V[4] ctxdebug realm ocm realm test
V[3] ctxinfo realm ocm realm test
V[2] ctxwarn realm ocm realm test
ERROR <nil> ctxerror realm ocm realm test
`))
	})

	It("sets log file", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-L", "logfile", "logtest")).To(Succeed())

		data, err := vfs.ReadFile(env.FileSystem(), "logfile")
		Expect(err).To(Succeed())

		fmt.Printf("%s\n", string(data))
		// 2024-06-16T13:59:34+02:00 warning [test] warn
		// 2024-06-16T13:59:34+02:00 error   [test] error
		// 2024-06-16T13:59:34+02:00 warning [test] ctxwarn
		// 2024-06-16T13:59:34+02:00 error   [test] ctxerror
		Expect(string(data)).To(MatchRegexp(`.* warning \[test\] warn
.* error   \[test\] error
.* warning \[test\] ctxwarn
.* error   \[test\] ctxerror
`))
	})

	It("sets attr from file", func() {
		buf := bytes.NewBuffer(nil)
		attr := mapocirepoattr.Get(env.Context)
		Expect(attr.Mode).To(Equal(mapocirepoattr.NoneMode))
		Expect(env.CatchOutput(buf).Execute("-X", "mapocirepo=@testdata/attr.yaml", "version")).To(Succeed())
		attr = mapocirepoattr.Get(env.Context)
		Expect(attr.Mode).To(Equal(mapocirepoattr.ShortHashMode))
	})
})
