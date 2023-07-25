// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/contexts/config"
	"github.com/open-component-model/ocm/v2/pkg/errors"
)

var _ = Describe("config handling", func() {

	var scheme config.ConfigTypeScheme
	var cfgctx config.Context

	BeforeEach(func() {
		scheme = config.NewConfigTypeScheme()
		cfgctx = config.WithConfigTypeScheme(scheme).New()
	})

	It("can deserialize unknown", func() {
		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		result, err := cfgctx.GetConfigForData(data, nil)
		Expect(err).To(Succeed())
		Expect(config.IsGeneric(result)).To(BeTrue())
	})

	It("can deserialize known", func() {
		RegisterAt(scheme)

		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		result, err := cfgctx.GetConfigForData(data, nil)
		Expect(err).To(Succeed())
		Expect(config.IsGeneric(result)).To(BeFalse())
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
		Expect(errors.IsErrUnknownKind(err, config.KIND_CONFIGTYPE)).To(BeTrue())
		Expect(config.IsGeneric(gen)).To(BeTrue())

		RegisterAt(scheme)
		d := newDummy(cfgctx)
		Expect(d.getApplied()).To(Equal([]*Config{cfg}))
	})

})
