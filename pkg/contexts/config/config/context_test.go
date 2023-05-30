// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	local "github.com/open-component-model/ocm/pkg/contexts/config/config"
)

var _ = Describe("generic config handling", func() {

	var scheme config.ConfigTypeScheme
	var cfgctx config.Context

	testdataconfig, _ := os.ReadFile("testdata/config.yaml")
	testdatajson, _ := yaml.YAMLToJSON(testdataconfig)

	_ = testdatajson

	BeforeEach(func() {
		scheme = config.NewConfigTypeScheme()
		scheme.AddKnownTypes(config.DefaultContext().ConfigTypes())
		cfgctx = config.WithConfigTypeScheme(scheme).New()
	})

	It("can deserialize config", func() {
		result, err := cfgctx.GetConfigForData(testdataconfig, nil)
		Expect(err).To(Succeed())
		Expect(config.IsGeneric(result)).To(BeFalse())
		Expect(reflect.TypeOf(result).String()).To(Equal("*config.Config"))
	})

	It("it applies to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg, err := cfgctx.GetConfigForData(testdataconfig, nil)
		Expect(err).To(Succeed())

		err = cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))
	})

	It("it applies to existing context", func() {

		cfg, err := cfgctx.GetConfigForData(testdataconfig, nil)
		Expect(err).To(Succeed())

		err = cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("testconfig: applying generic config list: {config entry 0--testconfig: config type \"Dummy\" is unknown, config entry 1--testconfig: config type \"Dummy\" is unknown}"))
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		RegisterAt(scheme)
		d := newDummy(cfgctx)
		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))
	})

	It("it applies composed config to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg := local.New()

		nested := NewConfig("alice", "")
		cfg.AddConfig(nested)
		nested = NewConfig("", "bob")
		cfg.AddConfig(nested)

		err := cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))
	})

	It("it applies composed config set to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg := local.New()

		nested := NewConfig("alice", "")
		cfg.AddConfigToSet("test", nested)
		nested = NewConfig("", "bob")
		cfg.AddConfig(nested)

		err := cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(2)))
		Expect(len(cfgs)).To(Equal(2))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("", "bob")}))

		err = cfgctx.ApplyConfigSet("test")
		Expect(err).To(Succeed())

		gen, cfgs = cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))
		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("", "bob"), NewConfig("alice", "")}))
	})
})
