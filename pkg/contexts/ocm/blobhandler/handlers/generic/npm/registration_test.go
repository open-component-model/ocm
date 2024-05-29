package npm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/npm"
	"github.com/open-component-model/ocm/pkg/registrations"
)

var _ = Describe("Config deserialization Test Environment", func() {

	It("deserializes string", func() {
		cfg := Must(registrations.DecodeConfig[npm.Config]("test"))
		Expect(cfg).To(Equal(&npm.Config{Url: "test"}))
	})

	It("deserializes struct", func() {
		cfg := Must(registrations.DecodeConfig[npm.Config](`{"url":"test"}`))
		Expect(cfg).To(Equal(&npm.Config{Url: "test"}))
	})
})
