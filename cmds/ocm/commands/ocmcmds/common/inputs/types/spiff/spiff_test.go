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

package spiff_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/spiff"
	"github.com/open-component-model/ocm/pkg/common"
)

var _ = Describe("spiff processing", func() {
	var env *TestEnv
	var ictx inputs.Context

	nv := common.NewNameVersion("test", "v1")
	BeforeEach(func() {
		env = NewTestEnv(TestData())
		ictx = inputs.NewContext(env.Context, common.NewPrinter(env.Context.StdOut()))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("processes template", func() {
		spec, err := spiff.New("test1.yaml", "", false, nil)
		Expect(err).To(Succeed())
		blob, s, err := spec.GetBlob(ictx, nv, "/testdata/dummy")
		Expect(err).To(Succeed())
		Expect(s).To(Equal(""))
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect("\n" + string(data)).To(Equal(`
alice: 24
bob: 25
`))
	})
	It("processes template with values", func() {
		spec, err := spiff.New("test1.yaml", "", false, map[string]interface{}{"diff": 2})
		Expect(err).To(Succeed())
		blob, s, err := spec.GetBlob(ictx, nv, "/testdata/dummy")
		Expect(err).To(Succeed())
		Expect(s).To(Equal(""))
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect("\n" + string(data)).To(Equal(`
alice: 24
bob: 26
`))
	})
	It("processes template with values with local working directory", func() {
		spec, err := spiff.New("test.yaml", "", false, map[string]interface{}{"diff": 2})
		Expect(err).To(Succeed())
		blob, s, err := spec.GetBlob(ictx, nv, "/testdata/subdir/dummy")
		Expect(err).To(Succeed())
		Expect(s).To(Equal(""))
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect("\n" + string(data)).To(Equal(`
alice: 24
bob: 26
`))

	})
})
