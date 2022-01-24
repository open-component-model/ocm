// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ocm_test

import (
	"encoding/json"

	"github.com/gardener/ocm/pkg/oci/repositories/empty"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
	ocmreg "github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
	"github.com/gardener/ocm/pkg/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = ocm.New()

var _ = Describe("access method", func() {
	It("instantiate repo mapped to empty oci repo", func() {
		backendSpec := genericocireg.NewGenericOCIBackendSpec(
			empty.NewRepositorySpec(),
			ocmreg.NewComponentRepositoryMeta(ocmreg.OCIRegistryDigestMapping))
		data, err := json.Marshal(backendSpec)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"componentNameMapping\":\"sha256-digest\",\"type\":\"Empty\"}"))

		repo, err := DefaultContext.RepositoryForConfig(data, runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(repo).NotTo(BeNil())
	})
})
