package v1_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "ocm.software/ocm/api/ocm/compdesc/equivalent/testhelper"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

var _ = Describe("types", func() {
	var labels v1.Labels
	var modtime *v1.Timestamp

	BeforeEach(func() {
		labels.Clear()
		labels.Set("label1", "value1", v1.WithSigning())
		labels.Set("label3", "value3")
	})

	Context("provider", func() {
		var prov *v1.Provider

		BeforeEach(func() {
			prov = &v1.Provider{
				Name:   "test",
				Labels: labels,
			}
		})

		It("detects equal", func() {
			CheckEquivalent(prov.Equivalent(*prov.Copy()))
		})

		It("detects name modification", func() {
			mod := prov.Copy()
			mod.Name = "mod"
			CheckNotLocalHashEqual(prov.Equivalent(*mod))
			CheckNotLocalHashEqual(mod.Equivalent(*prov))
		})

		It("detects volatile label modification", func() {
			mod := prov.Copy()
			mod.Labels.Set("label3", "mod")
			CheckNotEquivalent(prov.Equivalent(*mod))
			CheckNotEquivalent(mod.Equivalent(*prov))
		})

		It("detects non-volatile label modification", func() {
			mod := prov.Copy()
			mod.Labels.Set("label1", "mod")
			CheckNotLocalHashEqual(prov.Equivalent(*mod))
			CheckNotLocalHashEqual(mod.Equivalent(*prov))
		})
	})

	Context("object meta", func() {
		var meta *v1.ObjectMeta

		BeforeEach(func() {
			meta = &v1.ObjectMeta{
				Name:    "test",
				Version: "v1",
				Labels:  labels,
				Provider: v1.Provider{
					Name:   "provider",
					Labels: labels,
				},
				CreationTime: v1.NewTimestampP(),
			}
			modtime = v1.NewTimestampPFor(time.Now().Add(time.Second))
			_ = modtime
		})

		It("detects equal", func() {
			CheckEquivalent(meta.Equivalent(*meta.Copy()))
		})

		It("detects name modification", func() {
			mod := meta.Copy()
			mod.Name = "mod"
			CheckNotLocalHashEqual(meta.Equivalent(*mod))
			CheckNotLocalHashEqual(mod.Equivalent(*meta))
		})

		It("detects version modification", func() {
			mod := meta.Copy()
			mod.Version = "mod"
			CheckNotLocalHashEqual(meta.Equivalent(*mod))
			CheckNotLocalHashEqual(mod.Equivalent(*meta))
		})

		It("detects volatile provider modification", func() {
			mod := meta.Copy()
			mod.Provider.Labels.Set("label3", "mod")
			CheckNotEquivalent(meta.Equivalent(*mod))
			CheckNotEquivalent(mod.Equivalent(*meta))
		})

		It("detects non-volatile provider modification", func() {
			mod := meta.Copy()
			mod.Provider.Labels.Set("label1", "mod")
			CheckNotLocalHashEqual(meta.Equivalent(*mod))
			CheckNotLocalHashEqual(mod.Equivalent(*meta))
		})

		It("detects volatile label modification", func() {
			mod := meta.Copy()
			mod.Labels.Set("label3", "mod")
			CheckNotEquivalent(meta.Equivalent(*mod))
			CheckNotEquivalent(mod.Equivalent(*meta))
		})

		It("detects non-volatile label modification", func() {
			mod := meta.Copy()
			mod.Labels.Set("label1", "mod")
			CheckNotLocalHashEqual(meta.Equivalent(*mod))
			CheckNotLocalHashEqual(mod.Equivalent(*meta))
		})
	})
})
