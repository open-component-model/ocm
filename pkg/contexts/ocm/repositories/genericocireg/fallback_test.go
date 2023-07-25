// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package genericocireg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/v2/pkg/testutils"
)

var _ = Describe("decode fallback", func() {

	It("creates a dummy component", func() {
		specdata := `
type: other/v1
subPath: test
other: value
`
		spec := testutils.Must(DefaultContext.RepositoryTypes().Decode([]byte(specdata), nil))
		Expect(ocm.IsUnknownRepositorySpec(spec.(ocm.RepositorySpec))).To(BeTrue())
	})
})
