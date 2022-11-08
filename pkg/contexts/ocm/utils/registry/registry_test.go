// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/registry"
)

var amMey = registry.RegistrationKey{}.SetArtefact("a", "m")
var amtarKey = registry.RegistrationKey{}.SetArtefact("a", "m+tar")

var _ = Describe("lookup", func() {
	var reg *registry.Registry[string, registry.RegistrationKey]

	BeforeEach(func() {
		reg = registry.NewRegistry[string, registry.RegistrationKey]()
	})

	It("looks up complete", func() {
		reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m"), "test")
		reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
		reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

		h := reg.LookupHandler(amMey)
		Expect(h).To(Equal([]string{"test"}))
	})

	It("looks up partial artifact", func() {
		reg.Register(registry.RegistrationKey{}.SetArtefact("a", ""), "test")
		reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
		reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

		h := reg.LookupHandler(amMey)
		Expect(h).To(Equal([]string{"test"}))
	})

	It("looks up partial media", func() {
		reg.Register(registry.RegistrationKey{}.SetArtefact("", "m"), "test")
		reg.Register(registry.RegistrationKey{}.SetArtefact("a", "m1"), "testm")
		reg.Register(registry.RegistrationKey{}.SetArtefact("a1", "m"), "testa")

		h := reg.LookupHandler(amMey)
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
