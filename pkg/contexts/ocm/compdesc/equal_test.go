// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var _ = Describe("equivalence", func() {
	var labels v1.Labels
	var modtime *v1.Timestamp

	_ = modtime

	BeforeEach(func() {
		labels.Clear()
		labels.Set("label1", "value1", v1.WithSigning())
		labels.Set("label3", "value3")
	})

	Context("resources", func() {

	})

	Context("sources", func() {

	})

	Context("references", func() {

	})
})
