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

package download_test

import (
	"bytes"

	. "github.com/gardener/ocm/cmds/ocm/testhelper"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/mime"
	"github.com/gardener/ocm/pkg/oci/grammar"
	"github.com/gardener/ocm/pkg/oci/repositories/artefactset"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ARCH = "/tmp/ctf"
const VERSION = "v1"
const NS = "mandelsoft/test"
const OUT = "/tmp/res"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists single resource in ctf file", func() {
		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NS, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "artefact", "-O", OUT, "-r", ARCH, NS+grammar.RepositorySeparator+VERSION)).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(
			`
/tmp/res: downloaded
`))
		Expect(env.DirExists(OUT)).To(BeTrue())
		Expect(env.ReadFile(OUT + "/" + artefactset.ArtefactSetDescriptorFileName)).To(Equal([]byte("{\"schemaVersion\":2,\"mediaType\":\"application/vnd.oci.image.index.v1+json\",\"manifests\":[{\"mediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"digest\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\",\"size\":342,\"annotations\":{\"cloud.gardener.ocm/tags\":\"v1\"}}],\"annotations\":{\"cloud.gardener.ocm/main\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\"}}")))
	})
})
