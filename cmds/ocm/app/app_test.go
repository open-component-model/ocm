// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package app_test

import (
	"bytes"
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

var _ = Describe("Test Environment", func() {
	var log bytes.Buffer
	var env *TestEnv
	var oldlog *ocmlog.StaticContext

	BeforeEach(func() {
		oldlog = ocmlog.Context()
		log.Reset()
		def := buflogr.NewWithBuffer(&log)
		n := ocmlog.NewContext(logging.New(def))
		ocmlog.SetContext(n)
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
		ocmlog.SetContext(oldlog)
	})

	It("get gets the version", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("version")).To(Succeed())
		Expect(strings.HasPrefix(buf.String(), "version.Info{Major:")).To(BeTrue())

	})
	It("do logging", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
ERROR <nil> test error
ERROR <nil> test ctxerror
`))
	})

	It("sets logging", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-l", "Debug", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] test debug
V[3] test info
V[2] test warn
ERROR <nil> test error
V[4] test ctxdebug
V[3] test ctxinfo
V[2] test ctxwarn
ERROR <nil> test ctxerror
`))
	})

	It("sets logging by config", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "--logconfig", "testdata/logcfg.yaml", "logtest")).To(Succeed())
		Expect(log.String()).To(StringEqualTrimmedWithContext(`
V[4] test debug
V[3] test info
V[2] test warn
ERROR <nil> test error
V[4] test ctxdebug
V[3] test ctxinfo
V[2] test ctxwarn
ERROR <nil> test ctxerror
`))
	})

	It("sets log file", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).ExecuteModified(addTestCommands, "-L", "logfile", "logtest")).To(Succeed())

		data, err := vfs.ReadFile(env.FileSystem(), "logfile")
		Expect(err).To(Succeed())

		Expect(len(string(data))).To(Equal(183))
	})

})
