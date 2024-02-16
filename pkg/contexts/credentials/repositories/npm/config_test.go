package npm

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Config deserialization Test Environment", func() {
	It("read .npmrc", func() {
		cfg, err := readNpmConfigFile("testdata/.npmrc")
		Expect(err).To(BeNil())
		Expect(cfg).ToNot(BeNil())
		Expect(cfg).ToNot(BeEmpty())
		Expect(cfg["https://registry.npmjs.org/"]).To(Equal("npm_TOKEN"))
		Expect(cfg["https://npm.registry.acme.com/api/npm/"]).To(Equal("bearer_TOKEN"))
	})
})
