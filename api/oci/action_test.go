package oci_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/oci"
	oci_repository_prepare "ocm.software/ocm/api/oci/extensions/actions/oci-repository-prepare"
)

var _ = Describe("action registration", func() {
	It("registers oci prepare", func() {
		a := oci.DefaultContext().GetActions().GetActionTypes().GetAction(oci_repository_prepare.Type)
		Expect(a).NotTo(BeNil())
		v := a.GetVersion("v1")
		Expect(v).NotTo(BeNil())
	})
})
