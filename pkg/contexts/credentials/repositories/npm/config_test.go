package npm

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/npm"
	"github.com/open-component-model/ocm/pkg/registrations"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Config deserialization Test Environment", func() {

	It("deserializes string", func() {
		cfg := Must(registrations.DecodeConfig[npm.Config]("test"))
		Expect(cfg).To(Equal(&npm.Config{Url: "test"}))
	})

	It("deserializes struct", func() {
		cfg := Must(registrations.DecodeConfig[npm.Config](`{"Url":"test"}`))
		Expect(cfg).To(Equal(&npm.Config{Url: "test"}))
	})

	FIt("read .npmrc", func() {
		cfg, err := readNpmConfigFile("testdata/.npmrc")
		Expect(err).To(BeNil())
		Expect(cfg).ToNot(BeNil())
		Expect(cfg["https://npm.registry.acme.com/api/npm/"]).To(Equal("bearer_TOKEN"))
		Expect(cfg["https://registry.npm.org/"]).To(Equal("npm_TOKEN"))
	})
})
