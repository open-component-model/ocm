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

package config

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const ATTR_KEY = "test"

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{})
}

type AttributeType struct {
}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
A Test attribute.
`
}

type Attribute struct {
	Value string `json:"value"`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(*Attribute); !ok {
		return nil, fmt.Errorf("boolean required")
	}
	return marshaller.Marshal(v)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value Attribute
	err := unmarshaller.Unmarshal(data, &value)
	return &value, err
}

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("generic attributes", func() {
	attribute := &Attribute{"TEST"}
	var ctx config.Context

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	Context("applies", func() {

		It("applies later attribute config", func() {

			sub := credentials.WithConfigs(ctx).New()
			spec := New()
			Expect(spec.AddAttribute(ATTR_KEY, attribute)).To(Succeed())
			Expect(ctx.ApplyConfig(spec, "test")).To(Succeed())

			Expect(sub.GetAttributes().GetAttribute(ATTR_KEY, nil)).To(Equal(attribute))
		})

		It("applies earlier attribute config", func() {

			spec := New()
			Expect(spec.AddAttribute(ATTR_KEY, attribute)).To(Succeed())
			Expect(ctx.ApplyConfig(spec, "test")).To(Succeed())

			sub := credentials.WithConfigs(ctx).New()
			Expect(sub.GetAttributes().GetAttribute(ATTR_KEY, nil)).To(Equal(attribute))
		})
	})
})
