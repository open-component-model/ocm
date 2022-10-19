// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/empty"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	ocmreg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var DefaultContext = ocm.New()

var _ = Describe("access method", func() {
	It("instantiate repo mapped to empty oci repo", func() {
		backendSpec := genericocireg.NewRepositorySpec(
			empty.NewRepositorySpec(),
			ocmreg.NewComponentRepositoryMeta("", ocmreg.OCIRegistryDigestMapping))
		data, err := json.Marshal(backendSpec)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"componentNameMapping\":\"sha256-digest\",\"type\":\"Empty\"}"))

		repo, err := DefaultContext.RepositoryForConfig(data, runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(repo).NotTo(BeNil())
	})
})
