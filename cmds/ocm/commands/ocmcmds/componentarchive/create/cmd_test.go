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
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf/comparch"

	. "github.com/gardener/ocm/cmds/ocm/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "mandelsoft", "/tmp/ca",
			"l1=value", "l2={\"name\":\"value\"}")).To(Succeed())
		Expect(vfs.DirExists(env.FileSystem(), "/tmp/ca")).To(BeTrue())
		data, err := vfs.ReadFile(env.FileSystem(), "/tmp/ca/"+comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Name).To(Equal("test.de/x"))
		Expect(cd.Version).To(Equal("v1"))
		Expect(string(cd.Provider)).To(Equal("mandelsoft"))
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
