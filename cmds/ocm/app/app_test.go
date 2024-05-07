package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/tonglil/buflogr"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/mapocirepoattr"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/logging/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"
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
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "logtest")).To(Succeed())
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
		// {"level":"error","msg":"error","realm":"test","time":"2024-03-27 09:54:19"}
		// {"level":"error","msg":"ctxerror","realm":"test","time":"2024-03-27 09:54:19"}
		Expect(len(string(data))).To(Equal(312))
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
