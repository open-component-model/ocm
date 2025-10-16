package registry_test

import (
	"github.com/mandelsoft/goutils/set"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/ocmutils/registry"
)

var (
	aKey     = registry.RegistrationKey{}.SetArtifact("a", "")
	mKey     = registry.RegistrationKey{}.SetArtifact("", "m")
	amKey    = registry.RegistrationKey{}.SetArtifact("a", "m")
	a1mKey   = registry.RegistrationKey{}.SetArtifact("a1", "m")
	am1Key   = registry.RegistrationKey{}.SetArtifact("a", "m1")
	amtarKey = registry.RegistrationKey{}.SetArtifact("a", "m+tar")
)

var _ = Describe("lookup", func() {
	var reg *registry.Registry[string, registry.RegistrationKey]

	BeforeEach(func() {
		reg = registry.NewRegistry[string, registry.RegistrationKey]()
	})

	Context("lookup handler", func() {
		It("looks up complete", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			h := reg.LookupHandler(amKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks up partial artifact", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", ""), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			h := reg.LookupHandler(amKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks up partial media", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			h := reg.LookupHandler(amKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks complete with media sub type", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			h := reg.LookupHandler(amtarKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks partial with media sub type", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			h := reg.LookupHandler(amtarKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("prefers art", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("", "m"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", ""), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			h := reg.LookupHandler(amtarKey)
			Expect(h).To(Equal([]string{"test"}))
		})
	})

	Context("lookup keys", func() {
		It("fills missing", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			keys := reg.LookupKeys(aKey)
			Expect(keys).To(Equal(set.New(amKey, am1Key)))
		})

		It("fills missing", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m+tar"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			keys := reg.LookupKeys(mKey)
			Expect(keys).To(Equal(set.New(a1mKey)))
		})
		It("fills more specific media", func() {
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m+tar"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtifact("a1", "m"), "testa")

			keys := reg.LookupKeys(amKey)
			Expect(keys).To(Equal(set.New(amtarKey)))
		})
	})
})
