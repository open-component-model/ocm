// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logging_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/logging/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/logging"
	logcfg "github.com/mandelsoft/logging/config"
	"github.com/tonglil/buflogr"

	local "github.com/open-component-model/ocm/pkg/logging"
)

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("logging configuration", func() {
	var buf bytes.Buffer
	var ctx logging.Context

	BeforeEach(func() {
		local.SetContext(logging.NewDefault())
		buf.Reset()
		def := buflogr.NewWithBuffer(&buf)
		ctx = local.Context()
		ctx.SetBaseLogger(def)
	})

	It("just logs with defaults", func() {
		LogTest(ctx)
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
	})
	It("just logs with config", func() {
		r := logcfg.ConditionalRule("debug")
		cfg := &logcfg.Config{
			Rules: []logcfg.Rule{r},
		}

		Expect(local.Configure(cfg)).To(Succeed())
		LogTest(ctx)
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug realm ocm
V[3] info realm ocm
V[2] warn realm ocm
ERROR <nil> error realm ocm
`))
	})

})
