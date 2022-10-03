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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const CA = "/tmp/ca"
const CTF = "/tmp/ctf"
const VERSION = "v1"
const COMP = "test.de/x"
const COMP2 = "test.de/y"
const COMP3 = "test.de/z"
const PROVIDER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists single reference in component archive", func() {
		env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Reference("test", COMP2, VERSION)
			env.Reference("withid", COMP3, VERSION, func() {
				env.ExtraIdentity("id", "test")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "references", "-o", "wide", CA)).To(Succeed())
		ExpectTrimmedStringEqual(buf.String(),
			`
NAME   COMPONENT VERSION IDENTITY
test   test.de/y v1      "name"="test"
withid test.de/z v1      "id"="test","name"="withid"
`)
	})

	Context("with closure", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP2, VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("withid", COMP3, VERSION, func() {
						env.ExtraIdentity("id", "test")
					})
				})
				env.ComponentVersion(COMP3, VERSION, func() {
					env.Provider(PROVIDER)
				})
			})
			env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Reference("test", COMP2, VERSION)
			})
		})
		It("lists single reference in component archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "references", "--lookup", CTF, "-c", "-o", "wide", CA)).To(Succeed())
			ExpectTrimmedStringEqual(buf.String(),
				`
REFERENCEPATH              NAME   COMPONENT VERSION IDENTITY
test.de/x:v1               test   test.de/y v1      "name"="test"
test.de/x:v1->test.de/y:v1 withid test.de/z v1      "id"="test","name"="withid"
`)
		})
		It("lists flat tree in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "references", "-o", "tree", "--lookup", CTF, CA)).To(Succeed())
			ExpectTrimmedStringEqual(buf.String(),
				`
COMPONENTVERSION    NAME COMPONENT VERSION IDENTITY
└─ test.de/x:v1                            
   └─               test test.de/y v1      "name"="test"
`)
		})

		It("lits reference closure in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "references", "-c", "-o", "tree", "--lookup", CTF, CA)).To(Succeed())
			ExpectTrimmedStringEqual(buf.String(),
				`
COMPONENTVERSION    NAME   COMPONENT VERSION IDENTITY
└─ test.de/x:v1                              
   └─ ⊗             test   test.de/y v1      "name"="test"
      └─            withid test.de/z v1      "id"="test","name"="withid"
`)
		})
	})
})
