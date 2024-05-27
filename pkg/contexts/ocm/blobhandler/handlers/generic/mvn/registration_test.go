package mvn_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/mvn"
	"github.com/open-component-model/ocm/pkg/registrations"
)

var _ = Describe("Config deserialization Test Environment", func() {

	It("deserializes string", func() {
		cfg := Must(registrations.DecodeConfig[mvn.Config]("test"))
		Expect(cfg).To(Equal(&mvn.Config{Url: "test"}))
	})

	It("deserializes struct", func() {
		cfg := Must(registrations.DecodeConfig[mvn.Config](`{"url":"test"}`))
		Expect(cfg).To(Equal(&mvn.Config{Url: "test"}))
	})
})
