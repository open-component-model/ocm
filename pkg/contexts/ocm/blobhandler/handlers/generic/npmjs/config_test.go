package npmjs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/npmjs"
	"github.com/open-component-model/ocm/pkg/registrations"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Config deserialization Test Environment", func() {

	It("deserializes string", func() {
		cfg := Must(registrations.DecodeConfig[npmjs.Config]("test"))
		Expect(cfg).To(Equal(&npmjs.Config{"test"}))
	})

	It("deserializes struct", func() {
		cfg := Must(registrations.DecodeConfig[npmjs.Config](`{"Url":"test"}`))
		Expect(cfg).To(Equal(&npmjs.Config{"test"}))
	})

})
