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

package genericocireg_test

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
	ocmreg "github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
	"github.com/gardener/ocm/pkg/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultOCIContext = oci.NewDefaultContext(context.TODO())

var _ = Describe("access method", func() {
	specData := "{\"baseUrl\":\"X\",\"componentNameMapping\":\"sha256-digest\",\"type\":\"OCIRegistry\"}"

	It("marshal mapped spec", func() {
		gen := genericocireg.NewGenericOCIBackendSpec(
			ocireg.NewRepositorySpec("X"),
			ocmreg.NewComponentRepositoryMeta(ocmreg.OCIRegistryDigestMapping))
		data, err := json.Marshal(gen)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(specData))
	})

	It("decodes generic spec", func() {
		typ := genericocireg.NewOCIRepositoryBackendType(DefaultOCIContext)

		spec, err := typ.Decode([]byte(specData), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*genericocireg.GenericOCIBackendSpec"))

		eff, ok := spec.(*genericocireg.GenericOCIBackendSpec)
		Expect(ok).To(BeTrue())
		Expect(reflect.TypeOf(eff.RepositorySpec).String()).To(Equal("*ocireg.RepositorySpec"))
		Expect(eff.ComponentNameMapping).To(Equal(ocmreg.OCIRegistryDigestMapping))

		Expect(spec.GetType()).To(Equal(ocireg.OCIRegistryRepositoryType))
		effoci, ok := eff.RepositorySpec.(*ocireg.RepositorySpec)
		Expect(ok).To(BeTrue())
		Expect(effoci.BaseURL).To(Equal("X"))
	})
})
