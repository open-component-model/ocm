package runtime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("*** basic types", func() {
	Context("type name", func() {
		It("one arg", func() {
			t := runtime.TypeName("test")
			Expect(t).To(Equal("test"))
		})
		It("two arg", func() {
			t := runtime.TypeName("test", "v1")
			Expect(t).To(Equal("test" + runtime.VersionSeparator + "v1"))
		})
		It("two arg empty", func() {
			t := runtime.TypeName("test", "")
			Expect(t).To(Equal("test"))
		})
		It("two arg", func() {
			defer func() {
				e := recover()
				Expect(e).NotTo(BeNil())
			}()
			runtime.TypeName("test", "v1", "v3")
			Fail("no panic")
		})
	})
	Context("object type", func() {
		It("gets the type", func() {
			t := runtime.NewObjectType("test")
			Expect(t.GetType()).To(Equal("test"))
		})
		It("sets the type", func() {
			t := runtime.NewObjectType("test")
			t.SetType("other")
			Expect(t.GetType()).To(Equal("other"))
		})
	})

	Context("versioned object type", func() {
		It("get type and version of unversioned type", func() {
			t := runtime.NewVersionedTypedObject("test", "")
			Expect(t.GetType()).To(Equal("test"))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v1"))
		})
		It("get type and version of versioned type", func() {
			t := runtime.NewVersionedTypedObject("test", "v2")
			Expect(t.GetType()).To(Equal(runtime.TypeName("test", "v2")))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v2"))
		})

		It("set type", func() {
			t := runtime.NewVersionedTypedObject("test", "v2")
			t.SetType(runtime.TypeName("other", "v3"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("other", "v3")))
			Expect(t.GetKind()).To(Equal("other"))
			Expect(t.GetVersion()).To(Equal("v3"))
		})

		It("set kind on unversioned", func() {
			t := runtime.NewVersionedTypedObject("test")
			t.SetKind(runtime.TypeName("other"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("other")))
			Expect(t.GetKind()).To(Equal("other"))
			Expect(t.GetVersion()).To(Equal("v1"))
		})
		It("set version on unversioned", func() {
			t := runtime.NewVersionedTypedObject("test")
			t.SetVersion(runtime.TypeName("v3"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("test", "v3")))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v3"))
		})

		It("set kind on versioned", func() {
			t := runtime.NewVersionedTypedObject("test", "v2")
			t.SetKind(runtime.TypeName("other"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("other", "v2")))
			Expect(t.GetKind()).To(Equal("other"))
			Expect(t.GetVersion()).To(Equal("v2"))
		})
		It("set version on unversioned", func() {
			t := runtime.NewVersionedTypedObject("test", "v2")
			t.SetVersion(runtime.TypeName("v3"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("test", "v3")))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v3"))
		})
	})
})
