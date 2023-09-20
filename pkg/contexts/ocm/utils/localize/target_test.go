// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

var _ = Describe("value substitution in single target", func() {

	Context("localize", func() {
		const (
			ARCHIVE   = "archive.ctf"
			COMPONENT = "github.com/comp"
			VERSION   = "1.0.0"
			IMAGE     = "image"
		)

		var (
			repo ocm.Repository
			cv   ocm.ComponentVersionAccess
			env  *builder.Builder
		)

		BeforeEach(func() {
			env = builder.NewBuilder(nil)
			env.OCMCommonTransport(ARCHIVE, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider("mandelsoft")
						env.Resource(IMAGE, "", "Spiff", v1.LocalRelation, func() {
							env.ModificationOptions(ocm.SkipVerify())
							env.Digest("fake", "sha256", "fake")
							env.Access(ociartifact.New("ghcr.io/mandelsoft/test:v1"))
						})
					})
				})
			})

			var err error
			repo, err = ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, ARCHIVE, 0, env)
			Expect(err).To(Succeed())

			cv, err = repo.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
		})

		AfterEach(func() {
			Expect(cv.Close()).To(Succeed())
			Expect(repo.Close()).To(Succeed())
			vfs.Cleanup(env)
		})

		It("uses image ref data from component version", func() {
			mappings := UnmarshalImageMappings(`
- name: test1
  image: a.b.img
  resource:
    name: image
`)
			subst, err := localize.LocalizeMappings(mappings, cv, nil)
			Expect(err).To(Succeed())
			Expect(subst).To(Equal(UnmarshalValueMappings(`
- name: image mapping "test1"
  path: a.b.img
  value: ghcr.io/mandelsoft/test:v1
`)))
		})

		It("uses multiple resolved image ref data from component version", func() {

			mappings := UnmarshalImageMappings(`
- name: test1
  repository: a.b.rep
  tag: a.b.tag  
  image: a.b.img
  resource:
    name: image
`)
			subst, err := localize.LocalizeMappings(mappings, cv, nil)
			Expect(err).To(Succeed())
			Expect(subst).To(Equal(UnmarshalValueMappings(`
- name: image mapping "test1"-repository
  path: a.b.rep
  value: ghcr.io/mandelsoft/test
- name: image mapping "test1"-tag
  path: a.b.tag
  value: v1
- name: image mapping "test1"-image
  path: a.b.img
  value: ghcr.io/mandelsoft/test:v1
`)))
		})
	})

	Context("substitute", func() {
		var data = []byte(`
manifest:
  value1: orig1
  value2: orig2
`)

		It("handles simple values substitution", func() {
			subs := UnmarshalValueMappings(`
- name: test1
  path: manifest.value1
  value: config1
- name: test2
  path: manifest.value2
  value: config2
`)
			result, err := localize.SubstituteMappingsForData(subs, data)
			Expect(err).To(Succeed())

			Expect(string(result)).To(StringEqualTrimmedWithContext(`
manifest:
  value1: config1
  value2: config2
`))
		})

		It("handles json substitution", func() {
			subs := UnmarshalValueMappings(`
- name: test1
  path: manifest.value1
  value:
    some:
      value: 1
`)
			result, err := localize.SubstituteMappingsForData(subs, data)
			Expect(err).To(Succeed())

			Expect(string(result)).To(StringEqualTrimmedWithContext(`
manifest:
  value1:
      some:
        value: 1
  value2: orig2
`))
		})
	})
})
