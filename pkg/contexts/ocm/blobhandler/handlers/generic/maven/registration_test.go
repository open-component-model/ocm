package maven_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/maven"
	"github.com/open-component-model/ocm/pkg/registrations"
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
