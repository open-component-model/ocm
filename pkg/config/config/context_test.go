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
	"io/ioutil"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gardener/ocm/pkg/config"
	"sigs.k8s.io/yaml"
)

var _ = Describe("generic config handling", func() {

	var scheme config.ConfigTypeScheme
	var cfgctx config.Context

	testdataconfig, _ := ioutil.ReadFile("testdata/config.yaml")
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
		Expect(err.Error()).To(Equal("testconfig: {applying generic config list: config entry 0--testconfig: config type \"Dummy\" is unknown, config entry 1--testconfig: config type \"Dummy\" is unknown}"))
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		RegisterAt(scheme)
		d := newDummy(cfgctx)
		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))
	})

})
