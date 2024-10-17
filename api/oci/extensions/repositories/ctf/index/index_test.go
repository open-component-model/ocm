package index_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/oci/extensions/repositories/ctf/index"
)

var _ = Describe("index", func() {
	var rindex *RepositoryIndex

	BeforeEach(func() {
		rindex = NewRepositoryIndex()
	})

	Context("with digests only", func() {
		It("simple entry", func() {
			a1 := NewMeta("repo1", "", "digest1")
			rindex.AddArtifactInfo(a1)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a1,
			}))
		})
		It("two entries", func() {
			a1 := NewMeta("repo1", "", "digest1")
			a2 := NewMeta("repo1", "", "digest2")
			rindex.AddArtifactInfo(a1)
			rindex.AddArtifactInfo(a2)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtifactInfo("repo1", "digest2")).To(Equal(a2))
			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetArtifactInfos("digest2")).To(ContainElements(a2))

			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a1, *a2,
			}))
		})
	})
	Context("with tags", func() {
		It("simple entry", func() {
			a1 := NewMeta("repo1", "v1", "digest1")
			rindex.AddArtifactInfo(a1)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a1))

			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a1,
			}))
		})
		It("two entries two digests", func() {
			a1 := NewMeta("repo1", "v1", "digest1")
			a2 := NewMeta("repo1", "v2", "digest2")
			rindex.AddArtifactInfo(a1)
			rindex.AddArtifactInfo(a2)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a1))

			Expect(rindex.GetArtifactInfo("repo1", "digest2")).To(Equal(a2))
			Expect(rindex.GetArtifactInfo("repo1", "v2")).To(Equal(a2))

			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetArtifactInfos("digest2")).To(ContainElements(a2))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a1, *a2,
			}))
		})
		It("two tags one digest", func() {
			a1 := NewMeta("repo1", "v1", "digest1")
			a2 := NewMeta("repo1", "v2", "digest1")
			rindex.AddArtifactInfo(a1)
			rindex.AddArtifactInfo(a2)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a2))
			Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a1))

			Expect(rindex.GetArtifactInfo("repo1", "v2")).To(Equal(a2))

			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1, a2))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a1, *a2,
			}))
		})

		It("tag after digest", func() {
			a1 := NewMeta("repo1", "", "digest1")
			a2 := NewMeta("repo1", "v1", "digest1")
			rindex.AddArtifactInfo(a1)
			rindex.AddArtifactInfo(a2)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a2))
			Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a2))

			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a2))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a2,
			}))
		})

		It("reassign tag after digest", func() {
			a1 := NewMeta("repo1", "", "digest1")
			a2 := NewMeta("repo1", "v1", "digest1")
			a3 := NewMeta("repo1", "", "digest2")
			a4 := NewMeta("repo1", "v1", "digest2")
			rindex.AddArtifactInfo(a1)
			rindex.AddArtifactInfo(a2)
			rindex.AddArtifactInfo(a3)
			rindex.AddArtifactInfo(a4)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
			Expect(rindex.GetArtifactInfo("repo1", "digest2")).To(Equal(a4))
			Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a4))

			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1))
			Expect(rindex.GetArtifactInfos("digest2")).To(ContainElements(a4))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a1,
				*a4,
			}))
		})

		It("reassign tag after digest and second tag", func() {
			a1 := NewMeta("repo1", "", "digest1")
			a2 := NewMeta("repo1", "v1", "digest1")
			a2t := NewMeta("repo1", "v2", "digest1")
			a3 := NewMeta("repo1", "", "digest2")
			a4 := NewMeta("repo1", "v1", "digest2")
			rindex.AddArtifactInfo(a1)
			rindex.AddArtifactInfo(a2)
			rindex.AddArtifactInfo(a2t)
			rindex.AddArtifactInfo(a3)
			rindex.AddArtifactInfo(a4)

			Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a2t))
			Expect(rindex.GetArtifactInfo("repo1", "digest2")).To(Equal(a4))
			Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a4))

			Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a2t))
			Expect(rindex.GetArtifactInfos("digest2")).To(ContainElements(a4))
			Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
				*a4,
				*a2t,
			}))
		})

		Context("multiple repos", func() {
			It("simple entry", func() {
				a1 := NewMeta("repo1", "v1", "digest1")
				a2 := NewMeta("repo2", "v1", "digest2")
				rindex.AddArtifactInfo(a1)
				rindex.AddArtifactInfo(a2)

				Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
				Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a1))
				Expect(rindex.GetArtifactInfo("repo1", "digest2")).To(BeNil())

				Expect(rindex.GetArtifactInfo("repo2", "digest2")).To(Equal(a2))
				Expect(rindex.GetArtifactInfo("repo2", "v1")).To(Equal(a2))
				Expect(rindex.GetArtifactInfo("repo2", "digest1")).To(BeNil())

				Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1))
				Expect(rindex.GetArtifactInfos("digest2")).To(ContainElements(a2))
				Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
					*a1, *a2,
				}))
			})

			It("shared entry", func() {
				a1 := NewMeta("repo1", "v1", "digest1")
				a2 := NewMeta("repo2", "v2", "digest1")
				rindex.AddArtifactInfo(a1)
				rindex.AddArtifactInfo(a2)

				Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
				Expect(rindex.GetArtifactInfo("repo1", "v1")).To(Equal(a1))
				Expect(rindex.GetArtifactInfo("repo1", "v2")).To(BeNil())

				Expect(rindex.GetArtifactInfo("repo2", "digest1")).To(Equal(a2))
				Expect(rindex.GetArtifactInfo("repo2", "v2")).To(Equal(a2))
				Expect(rindex.GetArtifactInfo("repo2", "v1")).To(BeNil())

				Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1, a2))
				Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
					*a1, *a2,
				}))
			})

			It("shared entry without tag", func() {
				a1 := NewMeta("repo1", "", "digest1")
				a2 := NewMeta("repo2", "v2", "digest1")
				rindex.AddArtifactInfo(a1)
				rindex.AddArtifactInfo(a2)

				Expect(rindex.GetArtifactInfo("repo1", "digest1")).To(Equal(a1))
				Expect(rindex.GetArtifactInfo("repo1", "v2")).To(BeNil())

				Expect(rindex.GetArtifactInfo("repo2", "digest1")).To(Equal(a2))
				Expect(rindex.GetArtifactInfo("repo2", "v2")).To(Equal(a2))

				Expect(rindex.GetArtifactInfos("digest1")).To(ContainElements(a1, a2))
				Expect(rindex.GetDescriptor().Index).To(Equal([]ArtifactMeta{
					*a1, *a2,
				}))
			})
		})
	})
})
