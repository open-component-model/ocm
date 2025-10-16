package flagsets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

type Config = flagsets.Config

var (
	GetField = flagsets.GetField
	SetField = flagsets.SetField
)

var _ = Describe("config", func() {
	Context("get", func() {
		It("gets a field from empty map", func() {
			m := flagsets.Config{}
			Expect(GetField(m, "a")).To(BeNil())
		})

		It("gets a non existing field", func() {
			m := Config{}
			m["b"] = "vb"
			Expect(GetField(m, "a")).To(BeNil())
		})

		It("gets a flat field", func() {
			m := Config{}
			m["a"] = "va"
			Expect(GetField(m, "a")).To(Equal("va"))
		})

		It("gets a deep field", func() {
			m := Config{}
			a := Config{}
			m["a"] = a
			a["b"] = "vb"
			Expect(GetField(m, "a", "b")).To(Equal("vb"))
		})

		It("fails for non map", func() {
			m := Config{}
			m["a"] = "va"
			_, err := GetField(m, "a", "b")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("a is no map"))
		})

		It("fails for deep non map", func() {
			m := Config{}
			a := Config{}
			m["a"] = a
			a["b"] = "va"
			_, err := GetField(m, "a", "b", "c")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("a.b is no map"))
		})
	})

	Context("set", func() {
		It("fails for empty map", func() {
			err := SetField(nil, "x", "a")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("no map given"))
		})

		It("sets flat field", func() {
			m := Config{}
			err := SetField(m, "x", "a")
			Expect(err).To(Succeed())
			Expect(m).To(Equal(Config{"a": "x"}))
		})

		It("adds flat field", func() {
			m := Config{"b": "y"}
			err := SetField(m, "x", "a")
			Expect(err).To(Succeed())
			Expect(m).To(Equal(Config{"a": "x", "b": "y"}))
		})

		It("sets deep field", func() {
			m := Config{"a": Config{}}
			err := SetField(m, "x", "a", "b")
			Expect(err).To(Succeed())
			Expect(m).To(Equal(Config{"a": Config{"b": "x"}}))
		})

		It("inserts intermediate maps", func() {
			m := Config{"a": Config{}}
			err := SetField(m, "x", "a", "b", "c", "d")
			Expect(err).To(Succeed())
			Expect(m).To(Equal(Config{"a": Config{"b": Config{"c": Config{"d": "x"}}}}))
		})

		It("fails for non map", func() {
			m := Config{"a": "va"}
			err := SetField(m, "x", "a", "b")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("a is no map"))
		})

		It("fails for non map intermediate", func() {
			m := Config{"a": "va"}
			err := SetField(m, "x", "a", "b", "c")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("a is no map"))
		})
	})
})
