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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ctf"
const VERSION1 = "v1"
const VERSION2 = "v2"
const NS1 = "mandelsoft/test"
const NS2 = "mandelsoft/index"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NS1, func() {
				env.Manifest(VERSION1, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
				env.Manifest(VERSION2, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "otherdata")
					})
				})
			})

			env.Namespace(NS2, func() {
				env.Index(VERSION1, func() {
					env.Manifest("", func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
					env.Manifest("", func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "otherdata")
						})
					})
				})
				env.Manifest(VERSION2, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "yetanotherdata")
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get single artefacts", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "artefact", ARCH+"//"+NS1+":"+VERSION1)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
REGISTRY REPOSITORY      KIND     TAG DIGEST
         mandelsoft/test manifest v1  sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
`))
	})
	It("get all artefacts in namespace", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "artefact", ARCH+"//"+NS1)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
REGISTRY REPOSITORY      KIND     TAG DIGEST
         mandelsoft/test manifest v1  sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
         mandelsoft/test manifest v2  sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
`))
	})

	It("get all artefacts in other namespace", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "artefact", ARCH+"//"+NS2)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
REGISTRY REPOSITORY       KIND     TAG DIGEST
         mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
         mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
	})

	It("get closure of all artefacts in other namespace", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "artefact", "-c", ARCH+"//"+NS2)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
REFERENCEPATH                                                           REGISTRY REPOSITORY       KIND     TAG DIGEST
                                                                                 mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627          mandelsoft/index manifest -   sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627          mandelsoft/index manifest -   sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
                                                                                 mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
	})
	It("get tree of all tagged artefacts in other namespace", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "artefact", "-o", "tree", ARCH+"//"+NS2)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
NESTING REGISTRY REPOSITORY       KIND     TAG DIGEST
├─               mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
└─               mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
	})

	It("get tree of all artefacts in other namespace", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "artefact", "-c", "-o", "tree", ARCH+"//"+NS2)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
NESTING     REGISTRY REPOSITORY       KIND     TAG DIGEST
├─ ⊗                 mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
│  ├─                mandelsoft/index manifest -   sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
│  └─                mandelsoft/index manifest -   sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
└─                   mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
	})

})
