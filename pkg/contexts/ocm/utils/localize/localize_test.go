// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/common/accessobj"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/accessmethods/ociartifact"
	v1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/v2/pkg/env/builder"
)

var _ = Describe("image value mapping", func() {

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

		mappings := Localizations(`
- name: test1
  file: file1
  image: a.b.img
  resource:
    name: image
`)
		subst, err := localize.Localize(mappings, cv, nil)
		Expect(err).To(Succeed())
		Expect(subst).To(Equal(Substitutions(`
- name: image mapping "test1"
  file: file1
  path: a.b.img
  value: ghcr.io/mandelsoft/test:v1
`)))
	})

	It("uses multiple resolved image ref data from component version", func() {

		mappings := Localizations(`
- name: test1
  file: file1
  repository: a.b.rep
  tag: a.b.tag  
  image: a.b.img
  resource:
    name: image
`)
		subst, err := localize.Localize(mappings, cv, nil)
		Expect(err).To(Succeed())
		Expect(subst).To(Equal(Substitutions(`
- name: image mapping "test1"-repository
  file: file1
  path: a.b.rep
  value: ghcr.io/mandelsoft/test
- name: image mapping "test1"-tag
  file: file1
  path: a.b.tag
  value: v1
- name: image mapping "test1"-image
  file: file1
  path: a.b.img
  value: ghcr.io/mandelsoft/test:v1
`)))
	})
})
