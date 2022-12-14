// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("Input Type", func() {
	var env *InputTest

	BeforeEach(func() {
		env = NewInputTest(TYPE)
	})

	It("simple string decode", func() {
		env.Set(options.CompressOption, "true")
		env.Set(options.MediaTypeOption, "media")
		env.Set(options.DataOption, "!stringdata")
		env.Check(&Spec{
			Data:        runtime.Binary("stringdata"),
			ProcessSpec: cpi.NewProcessSpec("media", true),
		})
	})

	It("binary decode", func() {
		env.Set(options.CompressOption, "true")
		env.Set(options.MediaTypeOption, "media")
		env.Set(options.DataOption, "IXN0cmluZ2RhdGE=")
		env.Check(&Spec{
			Data:        runtime.Binary("!stringdata"),
			ProcessSpec: cpi.NewProcessSpec("media", true),
		})
	})
})
