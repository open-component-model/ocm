// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	"encoding/json"
	"github.com/open-component-model/ocm/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

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

	It("spec without key function", func() {
		spec := ocmreg.NewRepositorySpec("gcr.io", nil)
		key := Must(utils.Key(spec))
		Expect(key).To(Equal(string(Must(json.Marshal(spec)))))
	})

	It("spec with key function", func() {
		key := Must(utils.Key(&test_spec{}))
		Expect(key).To(Equal(TEST_KEY))
	})
})
