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

package create_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv3 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/ocm.gardener.cloud/v3alpha1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates comp arch", func() {

		plabels := metav1.Labels{}
		plabels.Set("email", "info@mandelsoft.de")
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "mandelsoft", "/tmp/ca",
			"l1=value", "l2={\"name\":\"value\"}", "-p", "email=info@mandelsoft.de")).To(Succeed())
		Expect(env.DirExists("/tmp/ca")).To(BeTrue())
		data, err := env.ReadFile("/tmp/ca/" + comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Name).To(Equal("test.de/x"))
		Expect(cd.Version).To(Equal("v1"))
		Expect(string(cd.Provider.Name)).To(Equal("mandelsoft"))
		Expect(cd.Provider.Labels).To(Equal(plabels))
		Expect(cd.Labels).To(Equal(metav1.Labels{
			{
				Name:  "l1",
				Value: []byte("\"value\""),
			},
			{
				Name:  "l2",
				Value: []byte("{\"name\":\"value\"}"),
			},
		}))
	})

	It("creates comp arch with "+compdescv3.SchemaVersion, func() {

		plabels := metav1.Labels{}
		plabels.Set("email", "info@mandelsoft.de")
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "mandelsoft", "/tmp/ca",
			"l1=value", "l2={\"name\":\"value\"}", "-p", "email=info@mandelsoft.de", "-S", compdescv3.SchemaVersion)).To(Succeed())
		Expect(env.DirExists("/tmp/ca")).To(BeTrue())
		data, err := env.ReadFile("/tmp/ca/" + comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Metadata.ConfiguredVersion).To(Equal(compdescv3.GroupVersion))
		Expect(cd.Name).To(Equal("test.de/x"))
		Expect(cd.Version).To(Equal("v1"))
		Expect(string(cd.Provider.Name)).To(Equal("mandelsoft"))
		Expect(cd.Provider.Labels).To(Equal(plabels))
		Expect(cd.Labels).To(Equal(metav1.Labels{
			{
				Name:  "l1",
				Value: []byte("\"value\""),
			},
			{
				Name:  "l2",
				Value: []byte("{\"name\":\"value\"}"),
			},
		}))
	})
})
