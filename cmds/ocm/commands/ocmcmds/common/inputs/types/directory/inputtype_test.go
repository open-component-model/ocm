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

package directory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
)

var _ = Describe("Input Type", func() {
	var env *InputTest

	var True = true
	var False = false

	BeforeEach(func() {
		env = NewInputTest(TYPE)
	})

	It("simple decode", func() {
		env.Set(options.PathOption, "mypath")
		env.Set(options.CompressOption, "true")
		env.Set(options.MediaTypeOption, "media")
		env.Set(options.PreserveDirOption, "false")
		env.Set(options.FollowSymlinksOption, "true")
		env.Set(options.IncludeOption, "x")
		env.Set(options.ExcludeOption, "a")
		env.Set(options.ExcludeOption, "b")
		env.Check(&Spec{
			MediaFileSpec: cpi.MediaFileSpec{
				PathSpec: cpi.PathSpec{
					Path: "mypath",
				},
				MediaType:        "media",
				CompressWithGzip: &True,
			},
			PreserveDir:    &False,
			IncludeFiles:   []string{"x"},
			ExcludeFiles:   []string{"a", "b"},
			FollowSymlinks: &True,
		})
	})

})
