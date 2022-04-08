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

package get_test

import (
	"bytes"

	. "github.com/gardener/ocm/cmds/ocm/testhelper"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/ocm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const COMP2 = "test.de/y"
const PROVIDER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	spec, err := ocm.NewGenericAccessSpec("{\"type\":\"git\"}")
	Expect(err).To(Succeed())

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists single resource in component archive", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Source("testdata", "v1", "git", func() {
				env.Access(spec)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "sources", ARCH)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
NAME     VERSION IDENTITY          TYPE
testdata v1      "name"="testdata" git
`))
	})

	It("lists single resource in ctf file", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Source("testdata", "v1", "git", func() {
						env.Access(spec)
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "sources", ARCH)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
NAME     VERSION IDENTITY          TYPE
testdata v1      "name"="testdata" git
`))
	})

	Context("with closure", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Source("testdata", "v1", "git", func() {
							env.Access(spec)
						})
					})
				})
				env.Component(COMP2, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Source("source", "v1", "git", func() {
							env.Access(spec)
						})
						env.Reference("base", COMP, VERSION)
					})
				})
			})
		})

		It("lists resource closure in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "sources", "-c", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect("\n" + buf.String()).To(Equal(
				`
REFERENCEPATH              NAME     VERSION IDENTITY          TYPE
test.de/y:v1               source   v1      "name"="source"   git
test.de/y:v1->test.de/x:v1 testdata v1      "name"="testdata" git
`))
		})
		It("lists flat tree in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "sources", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect("\n" + buf.String()).To(Equal(
				`
NESTING             NAME   VERSION IDENTITY        TYPE
└─ test.de/y:v1                                    
   └─               source v1      "name"="source" git
`))
		})

		It("lists resource closure in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "sources", "-c", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect("\n" + buf.String()).To(Equal(
				`
NESTING                NAME     VERSION IDENTITY          TYPE
└─ test.de/y:v1                                           
   ├─                  source   v1      "name"="source"   git
   └─ test.de/x:v1                                        
      └─               testdata v1      "name"="testdata" git
`))
		})
	})
})
