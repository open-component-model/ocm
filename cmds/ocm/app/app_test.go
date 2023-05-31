// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"bytes"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/tonglil/buflogr"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/logging/testhelper"
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
ERROR <nil> ocm/test error
ERROR <nil> ocm/test ctxerror
`))
	})

	It("sets logging", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-X", "plugindir=xxx", "-l", "Debug", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] ocm/test debug
V[3] ocm/test info
V[2] ocm/test warn
ERROR <nil> ocm/test error
V[4] ocm/test ctxdebug
V[3] ocm/test ctxinfo
V[2] ocm/test ctxwarn
ERROR <nil> ocm/test ctxerror
`))
	})

	It("sets logging by config", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "--logconfig", "@testdata/logcfg.yaml", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] ocm/test debug
V[3] ocm/test info
V[2] ocm/test warn
ERROR <nil> ocm/test error
V[4] ocm/test ctxdebug
V[3] ocm/test ctxinfo
V[2] ocm/test ctxwarn
ERROR <nil> ocm/test ctxerror
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
V[4] ocm/test debug
V[3] ocm/test info
V[2] ocm/test warn
ERROR <nil> ocm/test error
V[4] ocm/test ctxdebug
V[3] ocm/test ctxinfo
V[2] ocm/test ctxwarn
ERROR <nil> ocm/test ctxerror
`))
	})

	It("sets log file", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-L", "logfile", "logtest")).To(Succeed())

		data, err := vfs.ReadFile(env.FileSystem(), "logfile")
		Expect(err).To(Succeed())

		Expect(len(string(data))).To(Equal(191))
	})

})
