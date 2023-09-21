// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	. "github.com/open-component-model/ocm/pkg/testutils"

	. "github.com/onsi/ginkgo/v2"
	ocmreg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
)

var TEST_KEY = "test"

type test_spec struct {
	ocmreg.RepositorySpec
}

func (*test_spec) Key() (string, error) {
	return TEST_KEY, nil
}

var _ = Describe("session", func() {
	var session = ocm.NewSession(datacontext.NewSession())

	It("spec without key function", func() {
		spec := ocmreg.NewRepositorySpec("gcr.io", nil)
		key := Must(session.Key(spec))
		Expect(key).To(Equal(string(Must(json.Marshal(spec)))))
	})

	It("spec with key function", func() {
		key := Must(session.Key(&test_spec{}))
		Expect(key).To(Equal(TEST_KEY))
	})
})
