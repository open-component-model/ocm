// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/oci"
	oci_repository_prepare "github.com/open-component-model/ocm/pkg/contexts/oci/actions/oci-repository-prepare"
)

var _ = Describe("action registration", func() {
	It("registers oci prepare", func() {
		a := oci.DefaultContext().GetActions().GetActionTypes().GetAction(oci_repository_prepare.Type)
		Expect(a).NotTo(BeNil())
		v := a.GetVersion("v1")
		Expect(v).NotTo(BeNil())
	})
})
