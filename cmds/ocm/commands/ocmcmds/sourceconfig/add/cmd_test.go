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

	"github.com/open-component-model/ocm/pkg/testutils"
)

const SPECFILE = "/tmp/sources.yaml"
const VERSION = "v1"

func CheckSpec(env *TestEnv, spec string) {
	data, err := env.ReadFile(SPECFILE)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, string(data)).To(testutils.StringEqualTrimmedWithContext(spec))

}

var _ = Describe("Add sources", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("source by options", func() {
		It("adds simple text blob", func() {
			meta := `
name: testdata
type: PlainText
`
			input := `
type: file
path: ../testdata/testcontent
mediaType: text/plain
`
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--source", meta, "--input", input)).To(Succeed())
			CheckSpec(env, `
---
input:
  mediaType: text/plain
  path: ../testdata/testcontent
  type: file
name: testdata
type: PlainText
`)
		})

		It("defaults artifact type", func() {
			access := `
type: gitHub
repoUrl: github.com/open-component-model/ocm
commit: xxx
`
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--name", "sources", "--access", access)).To(Succeed())
			CheckSpec(env, `
---
access:
  commit: xxx
  repoUrl: github.com/open-component-model/ocm
  type: gitHub
name: sources
type: filesystem
`)
		})

		It("adds a second simple text blob", func() {
			meta1 := `
name: testdata1
type: PlainText
`
			meta2 := `
name: testdata2
type: PlainText
`
			input := `
type: file
path: ../testdata/testcontent
mediaType: text/plain
`
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--source", meta1, "--input", input)).To(Succeed())
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--source", meta2, "--input", input)).To(Succeed())
			CheckSpec(env, `
---
input:
  mediaType: text/plain
  path: ../testdata/testcontent
  type: file
name: testdata1
type: PlainText

---
input:
  mediaType: text/plain
  path: ../testdata/testcontent
  type: file
name: testdata2
type: PlainText
`)
		})
	})
})
