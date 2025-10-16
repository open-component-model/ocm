package maven_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/maven"
	"ocm.software/ocm/api/utils/registrations"
)

var _ = Describe("Config deserialization Test Environment", func() {
	It("deserializes string", func() {
		cfg := Must(registrations.DecodeConfig[maven.Config]("test"))
		Expect(cfg).To(Equal(&maven.Config{Url: "test"}))
	})

	It("deserializes struct", func() {
		cfg := Must(registrations.DecodeConfig[maven.Config](`{"url":"test"}`))
		Expect(cfg).To(Equal(&maven.Config{Url: "test"}))
	})
})
