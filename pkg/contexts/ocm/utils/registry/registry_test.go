// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/registry"
	"github.com/open-component-model/ocm/pkg/generics"
)

var aKey = registry.RegistrationKey{}.SetArtefact("a", "")
var mKey = registry.RegistrationKey{}.SetArtefact("", "m")
var amKey = registry.RegistrationKey{}.SetArtefact("a", "m")
var a1mKey = registry.RegistrationKey{}.SetArtefact("a1", "m")
var am1Key = registry.RegistrationKey{}.SetArtefact("a", "m1")
var amtarKey = registry.RegistrationKey{}.SetArtefact("a", "m+tar")

var _ = Describe("lookup", func() {
	var reg *registry.Registry[string, registry.RegistrationKey]

	BeforeEach(func() {
		reg = registry.NewRegistry[string, registry.RegistrationKey]()
	})

	Context("lookup handler", func() {
		It("looks up complete", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			h := reg.LookupHandler(amKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks up partial artefact", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", ""), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			h := reg.LookupHandler(amKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks up partial media", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			h := reg.LookupHandler(amKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks complete with media sub type", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			h := reg.LookupHandler(amtarKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("looks partial with media sub type", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			h := reg.LookupHandler(amtarKey)
			Expect(h).To(Equal([]string{"test"}))
		})

		It("prefers art", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("", "m"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", ""), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			h := reg.LookupHandler(amtarKey)
			Expect(h).To(Equal([]string{"test"}))
		})
	})

	Context("lookup keys", func() {
		It("fills missing", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			keys := reg.LookupKeys(aKey)
			Expect(keys).To(Equal(generics.NewSet(amKey, am1Key)))
		})

		It("fills missing", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m+tar"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			keys := reg.LookupKeys(mKey)
			Expect(keys).To(Equal(generics.NewSet(a1mKey)))
		})
		It("fills more specific media", func() {
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m+tar"), "test")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
			reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

			keys := reg.LookupKeys(amKey)
			Expect(keys).To(Equal(generics.NewSet(amtarKey)))
		})
	})
})
