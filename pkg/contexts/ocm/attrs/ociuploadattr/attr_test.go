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

package ociuploadattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("attribute", func() {
	var ctx ocm.Context
	var cfgctx config.Context

	attr := &me.Attribute{Ref: "ref"}

	BeforeEach(func() {
		cfgctx = config.WithSharedAttributes(datacontext.New(nil)).New()
		credctx := credentials.WithConfigs(cfgctx).New()
		ocictx := oci.WithCredentials(credctx).New()
		ctx = ocm.WithOCIRepositories(ocictx).New()
	})
	It("local setting", func() {
		Expect(me.Get(ctx)).To(BeNil())
		Expect(me.Set(ctx, attr)).To(Succeed())
		Expect(me.Get(ctx)).To(BeIdenticalTo(attr))
	})

	It("global setting", func() {
		Expect(me.Get(cfgctx)).To(BeNil())
		Expect(me.Set(ctx, attr)).To(Succeed())
		Expect(me.Get(ctx)).To(BeIdenticalTo(attr))
	})

	It("parses string", func() {
		Expect(me.AttributeType{}.Decode([]byte("ref"), runtime.DefaultJSONEncoding)).To(Equal(&me.Attribute{Ref: "ref"}))
	})

	It("parses spec", func() {
		spec, err := oci.ToGenericRepositorySpec(ocireg.NewRepositorySpec("ghcr.io"))
		Expect(err).To(Succeed())
		attr := &me.Attribute{
			Repository:      spec,
			NamespacePrefix: "ref",
		}
		data, err := json.Marshal(attr)
		Expect(err).To(Succeed())
		Expect(me.AttributeType{}.Decode(data, runtime.DefaultJSONEncoding)).To(Equal(attr))
	})
})
