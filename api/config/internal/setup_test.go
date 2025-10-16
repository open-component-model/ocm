package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/config/internal"
)

var _ = Describe("setup", func() {
	It("creates initial", func() {
		Expect(len(config.DefaultContext().ConfigTypes().KnownTypeNames())).To(Equal(6))
		Expect(len(internal.DefaultConfigTypeScheme.KnownTypeNames())).To(Equal(6))
	})
})
