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

package config_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	config2 "github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/errors"
)

var _ = Describe("config handling", func() {

	var scheme config2.ConfigTypeScheme
	var cfgctx config2.Context

	BeforeEach(func() {
		scheme = config2.NewConfigTypeScheme()
		cfgctx = config2.WithConfigTypeScheme(scheme).New()
	})

	It("can deserialize unknown", func() {
		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		result, err := cfgctx.GetConfigForData(data, nil)
		Expect(err).To(Succeed())
		Expect(config2.IsGeneric(result)).To(BeTrue())
	})

	It("can deserialize known", func() {
		RegisterAt(scheme)

		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		result, err := cfgctx.GetConfigForData(data, nil)
		Expect(err).To(Succeed())
		Expect(config2.IsGeneric(result)).To(BeFalse())
		Expect(reflect.TypeOf(result).String()).To(Equal("*config_test.Config"))
	})

	It("it applies to existing context", func() {
		RegisterAt(scheme)

		d := newDummy(cfgctx)

		cfg := NewConfig("a", "b")

		err := cfgctx.ApplyConfig(cfg, "test")

		Expect(err).To(Succeed())

		Expect(d.getApplied()).To(Equal([]*Config{cfg}))
	})

	It("it applies to new context", func() {
		RegisterAt(scheme)

		cfg := NewConfig("a", "b")

		err := cfgctx.ApplyConfig(cfg, "test")
		Expect(err).To(Succeed())

		d := newDummy(cfgctx)
		Expect(d.applied).To(Equal([]*Config{cfg}))
	})

	It("it applies generic to new context", func() {

		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		gen, err := cfgctx.ApplyData(data, nil, "test")
		Expect(err).To(HaveOccurred())
		Expect(errors.IsErrUnknownKind(err, config2.KIND_CONFIGTYPE)).To(BeTrue())
		Expect(config2.IsGeneric(gen)).To(BeTrue())

		RegisterAt(scheme)
		d := newDummy(cfgctx)
		Expect(d.getApplied()).To(Equal([]*Config{cfg}))
	})

})
