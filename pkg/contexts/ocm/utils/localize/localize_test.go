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
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

var _ = Describe("image value mapping", func() {

	const ARCHIVE = "archive.ctf"
	const COMPONENT = "github.com/comp"
	const VERSION = "1.0.0"
	const IMAGE = "image"

	var repo ocm.Repository
	var cv ocm.ComponentVersionAccess
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder(nil)
		env.OCMCommonTransport(ARCHIVE, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider("mandelsoft")
					env.Resource(IMAGE, "", "Spiff", v1.LocalRelation, func() {
						env.Access(ociartefact.New("ghcr.io/mandelsoft/test:v1"))
					})
				})
			})
		})

		var err error
		repo, err = ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, ARCHIVE, 0, env)
		Expect(err).To(Succeed())

		cv, err = repo.LookupComponentVersion(COMPONENT, VERSION)
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
- name: test1
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
- name: test1-repository
  file: file1
  path: a.b.rep
  value: ghcr.io/mandelsoft/test
- name: test1-tag
  file: file1
  path: a.b.tag
  value: v1
- name: test1-image
  file: file1
  path: a.b.img
  value: ghcr.io/mandelsoft/test:v1
`)))
	})
})
