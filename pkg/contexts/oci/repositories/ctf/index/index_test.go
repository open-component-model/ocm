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

package index_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/index"
)

var _ = Describe("index", func() {
	var rindex *RepositoryIndex

	BeforeEach(func() {
		rindex = NewRepositoryIndex()
	})

	Context("with digests only", func() {
		It("simple entry", func() {
			a1 := NewMeta("repo1", "", "digest1")
			rindex.AddArtefactInfo(a1)

			Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
				*a1,
			}))
		})
		It("two entries", func() {
			a1 := NewMeta("repo1", "", "digest1")
			a2 := NewMeta("repo1", "", "digest2")
			rindex.AddArtefactInfo(a1)
			rindex.AddArtefactInfo(a2)

			Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtefactInfo("repo1", "digest2")).To(Equal(a2))
			Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetArtefactInfos("digest2")).To(ContainElements(a2))

			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
				*a1, *a2,
			}))
		})
	})
	Context("with tags", func() {
		It("simple entry", func() {
			a1 := NewMeta("repo1", "v1", "digest1")
			rindex.AddArtefactInfo(a1)

			Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtefactInfo("repo1", "v1")).To(Equal(a1))

			Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
				*a1,
			}))
		})
		It("two entries two digests", func() {
			a1 := NewMeta("repo1", "v1", "digest1")
			a2 := NewMeta("repo1", "v2", "digest2")
			rindex.AddArtefactInfo(a1)
			rindex.AddArtefactInfo(a2)

			Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtefactInfo("repo1", "v1")).To(Equal(a1))

			Expect(rindex.GetArtefactInfo("repo1", "digest2")).To(Equal(a2))
			Expect(rindex.GetArtefactInfo("repo1", "v2")).To(Equal(a2))

			Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetArtefactInfos("digest2")).To(ContainElements(a2))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
				*a1, *a2,
			}))
		})
		It("two tags one digest", func() {
			a1 := NewMeta("repo1", "v1", "digest1")
			a2 := NewMeta("repo1", "v2", "digest1")
			rindex.AddArtefactInfo(a1)
			rindex.AddArtefactInfo(a2)

			Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a2))
			Expect(rindex.GetArtefactInfo("repo1", "v1")).To(Equal(a1))

			Expect(rindex.GetArtefactInfo("repo1", "v2")).To(Equal(a2))

			Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1, a2))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
				*a1, *a2,
			}))
		})

		It("tag after digest", func() {
			a1 := NewMeta("repo1", "", "digest1")
			a2 := NewMeta("repo1", "v1", "digest1")
			rindex.AddArtefactInfo(a1)
			rindex.AddArtefactInfo(a2)

			Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a2))
			Expect(rindex.GetArtefactInfo("repo1", "v1")).To(Equal(a2))

			Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a2))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
				*a2,
			}))
		})

		Context("multiple repos", func() {
			It("simple entry", func() {
				a1 := NewMeta("repo1", "v1", "digest1")
				a2 := NewMeta("repo2", "v1", "digest2")
				rindex.AddArtefactInfo(a1)
				rindex.AddArtefactInfo(a2)

				Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
				Expect(rindex.GetArtefactInfo("repo1", "v1")).To(Equal(a1))
				Expect(rindex.GetArtefactInfo("repo1", "digest2")).To(BeNil())

				Expect(rindex.GetArtefactInfo("repo2", "digest2")).To(Equal(a2))
				Expect(rindex.GetArtefactInfo("repo2", "v1")).To(Equal(a2))
				Expect(rindex.GetArtefactInfo("repo2", "digest1")).To(BeNil())

				Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1))
				Expect(rindex.GetArtefactInfos("digest2")).To(ContainElements(a2))
				Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
					*a1, *a2,
				}))
			})

			It("shared entry", func() {
				a1 := NewMeta("repo1", "v1", "digest1")
				a2 := NewMeta("repo2", "v2", "digest1")
				rindex.AddArtefactInfo(a1)
				rindex.AddArtefactInfo(a2)

				Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
				Expect(rindex.GetArtefactInfo("repo1", "v1")).To(Equal(a1))
				Expect(rindex.GetArtefactInfo("repo1", "v2")).To(BeNil())

				Expect(rindex.GetArtefactInfo("repo2", "digest1")).To(Equal(a2))
				Expect(rindex.GetArtefactInfo("repo2", "v2")).To(Equal(a2))
				Expect(rindex.GetArtefactInfo("repo2", "v1")).To(BeNil())

				Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1, a2))
				Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
					*a1, *a2,
				}))
			})

			It("shared entry without tag", func() {
				a1 := NewMeta("repo1", "", "digest1")
				a2 := NewMeta("repo2", "v2", "digest1")
				rindex.AddArtefactInfo(a1)
				rindex.AddArtefactInfo(a2)

				Expect(rindex.GetArtefactInfo("repo1", "digest1")).To(Equal(a1))
				Expect(rindex.GetArtefactInfo("repo1", "v2")).To(BeNil())

				Expect(rindex.GetArtefactInfo("repo2", "digest1")).To(Equal(a2))
				Expect(rindex.GetArtefactInfo("repo2", "v2")).To(Equal(a2))

				Expect(rindex.GetArtefactInfos("digest1")).To(ContainElements(a1, a2))
				Expect(rindex.GetDescriptor().Index).To(Equal([]ArtefactMeta{
					*a1, *a2,
				}))
			})
		})
	})
})
