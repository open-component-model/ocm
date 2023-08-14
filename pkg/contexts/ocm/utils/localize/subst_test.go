// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	env2 "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

var _ = Describe("value substitution in filesystem", func() {
	var (
		env       *builder.Builder
		payloadfs vfs.FileSystem
	)

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
		subs := UnmarshalSubstitutions(`
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
		subs := UnmarshalSubstitutions(`
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

	It("handles json substitution", func() {
		subs := UnmarshalSubstitutions(`
- name: test1
  file: dir/some.json
  path: manifest.value1
  value:
    some:
      value: 1
`)
		err := localize.Substitute(subs, payloadfs)
		Expect(err).To(Succeed())

		CheckFile("dir/some.json", payloadfs, `
{"manifest": {"value1": {"some": {"value": 1}}, "value2": "orig2"}}

`)
	})

})
