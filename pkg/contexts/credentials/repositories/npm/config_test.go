package npm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	npm2 "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/npm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/npm"
	"github.com/open-component-model/ocm/pkg/registrations"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Config deserialization Test Environment", func() {

	It("deserializes string", func() {
		cfg := Must(registrations.DecodeConfig[npm.Config]("test"))
		Expect(cfg).To(Equal(&npm.Config{"test"}))
	})

	It("deserializes struct", func() {
		cfg := Must(registrations.DecodeConfig[npm.Config](`{"Url":"test"}`))
		Expect(cfg).To(Equal(&npm.Config{"test"}))
	})

	It("read .npmrc", func() {
		cfg, err := npm2.ReadNpmConfigFile("testdata/.npmrc")
		Expect(err).To(BeNil())
		Expect(cfg).ToNot(BeNil())
		Expect(cfg).To(HaveKeyWithValue("npm.registry.acme.com/api/npm", "bearer_TOKEN"))
		Expect(cfg).To(HaveKeyWithValue("registry.npm.org", "npm_TOKEN"))
	})
})
