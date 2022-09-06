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

package add_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
)

const ARCH = "/tmp/ca"
const VERSION = "v1.1.1"
const REF = "github.com/mandelsoft/ref"

func CheckReference(env *TestEnv, cd *compdesc.ComponentDescriptor, name string) {
	r, err := cd.GetComponentReferenceByIdentity(metav1.NewIdentity(name))
	Expect(err).To(Succeed())
	Expect(r.Version).To(Equal(VERSION))
	Expect(r.ComponentName).To(Equal(REF))
}

var _ = Describe("Add references", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "mandelsoft", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds simple ref", func() {
		Expect(env.Execute("add", "references", ARCH, "/testdata/references.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds simple ref by cli env file", func() {
		Expect(env.Execute("add", "references", ARCH, "--settings", "/testdata/settings", "/testdata/references.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds simple ref by cli variable", func() {
		Expect(env.Execute("add", "references", ARCH, "VERSION=v1.1.1", "/testdata/references.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

})
