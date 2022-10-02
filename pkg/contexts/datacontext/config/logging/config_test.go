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

package logging_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/logging/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/logging"
	"github.com/tonglil/buflogr"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	logcfg "github.com/open-component-model/ocm/pkg/contexts/datacontext/config/logging"
	log "github.com/open-component-model/ocm/pkg/logging"
)

var _ = Describe("logging configuration", func() {
	var ctx datacontext.AttributesContext
	var cfg config.Context
	var buf bytes.Buffer

	BeforeEach(func() {
		logging.SetDefaultContext(logging.NewDefault())
		ctx = datacontext.New(nil)
		cfg = config.WithSharedAttributes(ctx).New()

		buf.Reset()
		def := buflogr.NewWithBuffer(&buf)
		ctx.LoggingContext().SetBaseLogger(def)
	})

	_ = cfg

	It("just logs with defaults", func() {
		LogTest(ctx)

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[3] info
V[2] warn
ERROR <nil> error
`))
	})

	It("just logs with settings from default context", func() {
		logging.DefaultContext().AddRule(logging.NewConditionRule(logging.DebugLevel))
		LogTest(ctx)

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
`))
	})

	It("just logs with settings from default context", func() {
		logging.DefaultContext().AddRule(logging.NewConditionRule(logging.DebugLevel))
		LogTest(cfg)

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
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
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
V[4] cfgdebug
V[3] cfginfo
V[2] cfgwarn
ERROR <nil> cfgerror
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
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
V[4] cfgdebug
V[3] cfginfo
V[2] cfgwarn
ERROR <nil> cfgerror
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
V[3] info
V[2] warn
ERROR <nil> error
V[4] cfgdebug
V[3] cfginfo
V[2] cfgwarn
ERROR <nil> cfgerror
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
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
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
V[3] info
V[2] warn
ERROR <nil> error
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
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
`))
		})
	})
})
