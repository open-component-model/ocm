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

package common_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
)

var _ = Describe("Blob Inputs", func() {

	It("missing input", func() {
		in := `
access:
  type: localBlob
`
		_, err := common.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})

	It("simple decode", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
`
		_, err := common.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})
	It("complains about additional input field", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
  bla: blub
`
		_, err := common.DecodeInput([]byte(in), nil)
		Expect(err.Error()).To(Equal("input.bla: Forbidden: unknown field"))
	})

	It("does not complains about additional dir field", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: dir
  excludeFiles:
     - xyz
`
		_, err := common.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})

	It("complains about additional dir field for file", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
  excludeFiles:
  - xyz
`
		_, err := common.DecodeInput([]byte(in), nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("input.excludeFiles: Forbidden: unknown field"))
	})
})
