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
	"encoding/json"

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
V[3] info
V[2] warn
ERROR <nil> error
`))
	})
	It("just logs with config", func() {
		data := `{ "rule": { "level": "Debug" } }`
		cfg := &logcfg.Config{
			Rules: []json.RawMessage{
				[]byte(data),
			},
		}

		Expect(local.Configure(cfg)).To(Succeed())
		LogTest(ctx)
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
V[4] debug
V[3] info
V[2] warn
ERROR <nil> error
`))
	})

})
