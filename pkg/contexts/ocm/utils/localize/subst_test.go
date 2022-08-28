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

package localize_test

import (
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	env2 "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

var _ = Describe("value substitution in filesystem", func() {
	var env *builder.Builder
	var payloadfs vfs.FileSystem
	BeforeEach(func() {
		env = builder.NewBuilder(env2.NewEnvironment(env2.ModifiableTestData()))
		fs, err := projectionfs.New(env.FileSystem(), "testdata")
		Expect(err).To(Succeed())
		payloadfs = fs
	})

	AfterEach(func() {
		vfs.Cleanup(payloadfs)
		vfs.Cleanup(env)
	})

	It("handles simple values substitution", func() {
		subs := Substitutions(`
- name: test1
  file: dir/manifest1.yaml
  path: manifest.value1
  value: config1
- name: test2
  file: dir/manifest2.yaml
  path: manifest.value2
  value: config2
`)
		err := localize.Substitute(subs, payloadfs)
		Expect(err).To(Succeed())

		CheckFile("dir/manifest1.yaml", payloadfs, `
manifest:
  value1: config1
  value2: orig2
`)
		CheckFile("dir/manifest2.yaml", payloadfs, `
manifest:
  value1: orig1
  value2: config2
`)
	})

	It("handles multiple values substitution", func() {
		subs := Substitutions(`
- name: test1
  file: dir/manifest1.yaml
  path: manifest.value1
  value: config1
- name: test2
  file: dir/manifest1.yaml
  path: manifest.value2
  value: config2
`)
		err := localize.Substitute(subs, payloadfs)
		Expect(err).To(Succeed())

		CheckFile("dir/manifest1.yaml", payloadfs, `
manifest:
  value1: config1
  value2: config2
`)
	})

})
