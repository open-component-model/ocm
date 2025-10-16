package mapocirepoattr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext"
	me "ocm.software/ocm/api/ocm/extensions/attrs/mapocirepoattr"
)

var _ = Describe("attribute", func() {
	var ctx datacontext.Context

	BeforeEach(func() {
		ctx = datacontext.New(nil)
	})

	It("set bool", func() {
		Expect(me.Get(ctx)).To(Equal(&me.Attribute{Mode: me.NoneMode}))
		ctx.GetAttributes().SetAttribute(me.ATTR_KEY, true)
		a := me.Get(ctx)
		Expect(a).To(Equal(&me.Attribute{Mode: me.ShortHashMode}))
		hash := "5afa3f0f1b63d64422e7f93e2d9792b7c1f3b4462a931d80b25703f7e6fc79c2"
		Expect(a.Map("very-long-path/with-many-path-segments/and-really-longer-than-a-hash/artifact")).To(Equal(hash[:8] + "/artifact"))
	})

	It("set attr", func() {
		ctx.GetAttributes().SetAttribute(me.ATTR_KEY, &me.Attribute{Mode: me.MappingMode, PrefixMappings: map[string]string{"a": "b", "a/b": "c"}})
		a := me.Get(ctx)
		Expect(a).To(Equal(&me.Attribute{Mode: me.MappingMode, PrefixMappings: map[string]string{"a": "b", "a/b": "c"}}))
		Expect(a.Map("a/b/c")).To(Equal("c/c"))
		Expect(a.Map("a/c")).To(Equal("b/c"))
		Expect(a.Map("x/y")).To(Equal("x/y"))
	})
})
