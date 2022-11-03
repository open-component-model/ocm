// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
)

var _ = Describe("lookup", func() {
	var reg *internal.Registry[string]

	BeforeEach(func() {
		reg = internal.NewRegistry[string]()
	})

	It("looks up complete", func() {
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m"}, "test")
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m1"}, "testm")
		reg.Register(internal.UploaderKey{ArtifactType: "a1", MediaType: "m"}, "testa")

		h, ok := reg.LookupHandler("a", "m")
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal("test"))
	})

	It("looks up partial artifact", func() {
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: ""}, "test")
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m1"}, "testm")
		reg.Register(internal.UploaderKey{ArtifactType: "a1", MediaType: "m"}, "testa")

		h, ok := reg.LookupHandler("a", "m")
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal("test"))
	})

	It("looks up partial media", func() {
		reg.Register(internal.UploaderKey{ArtifactType: "", MediaType: "m"}, "test")
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m1"}, "testm")
		reg.Register(internal.UploaderKey{ArtifactType: "a1", MediaType: "m"}, "testa")

		h, ok := reg.LookupHandler("a", "m")
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal("test"))
	})

	It("looks complete with media sub type", func() {
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m"}, "test")
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m1"}, "testm")
		reg.Register(internal.UploaderKey{ArtifactType: "a1", MediaType: "m"}, "testa")

		h, ok := reg.LookupHandler("a", "m+tar")
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal("test"))
	})

	It("looks partial with media sub type", func() {
		reg.Register(internal.UploaderKey{ArtifactType: "", MediaType: "m"}, "test")
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: "m1"}, "testm")
		reg.Register(internal.UploaderKey{ArtifactType: "a1", MediaType: "m"}, "testa")

		h, ok := reg.LookupHandler("a", "m+tar")
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal("test"))
	})

	It("prefers art", func() {
		reg.Register(internal.UploaderKey{ArtifactType: "", MediaType: "m"}, "testm")
		reg.Register(internal.UploaderKey{ArtifactType: "a", MediaType: ""}, "test")
		reg.Register(internal.UploaderKey{ArtifactType: "a1", MediaType: "m"}, "testa")

		h, ok := reg.LookupHandler("a", "m+tar")
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal("test"))
	})
})
