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
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gardener/ocm/pkg/oci/repositories/empty"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
	"github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
	"github.com/gardener/ocm/pkg/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func InOut(in runtime.TypedObject, encoding runtime.Encoding) (runtime.TypedObject, error) {
	out := reflect.New(reflect.TypeOf(in).Elem()).Interface().(runtime.TypedObject)
	data, err := encoding.Marshal(in)
	if err != nil {
		return nil, err
	}
	fmt.Printf("inout: %s\n", data)
	err = encoding.Unmarshal(data, out)
	return out, err
}


var DefaultContext = ocm.NewDefaultContext(context.TODO())

var _ = Describe("access method", func() {
	It("unmarshal json test", func() {
		out, err := InOut(runtime.NewEmptyUnstructured("test"), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))

		out, err = InOut(runtime.NewEmptyUnstructuredVersioned("test"), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))

		out, err = InOut(&core.UnknownRepositorySpec{*runtime.NewEmptyUnstructuredVersioned("test")}, runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
	})

	It("unmarshal yaml test", func() {
		out, err := InOut(runtime.NewEmptyUnstructured("test"), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))

		out, err = InOut(runtime.NewEmptyUnstructuredVersioned("test"), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))

		out, err = InOut(&core.UnknownRepositorySpec{*runtime.NewEmptyUnstructuredVersioned("test")}, runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
	})

	It("instantiate local blob access method for component archive", func() {
		backendSpec := genericocireg.NewGenericOCIBackendSpec(
			empty.NewEmptyRepositorySpec(),
			&ocireg.ComponentRepositoryMeta{
				ComponentNameMapping: ocireg.OCIRegistryDigestMapping,
			})
		data, err := json.Marshal(backendSpec)
		data2, err := json.Marshal(backendSpec.ComponentRepositoryMeta)

		other:=ocireg.NewOCIRegistryRepositorySpec("X", ocireg.OCIRegistryDigestMapping)
		data3, err := json.Marshal(other)

		Expect(err).To(Succeed())
		Expect(data).NotTo(BeNil())

		fmt.Printf("*** spec: %s\n", string(data))
		fmt.Printf("*** data2: %s\n", string(data2))
		fmt.Printf("*** data3: %s\n", string(data3))
		repo, err := DefaultContext.RepositoryForConfig(data, runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(repo).NotTo(BeNil())
	})
})
