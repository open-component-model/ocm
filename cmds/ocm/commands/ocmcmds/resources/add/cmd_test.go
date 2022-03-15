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
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf/comparch"

	. "github.com/gardener/ocm/cmds/ocm/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "mandelsoft", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds simple text blob", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/resources.yaml")).To(Succeed())
		data, err := vfs.ReadFile(env.FileSystem(), vfs.Join(nil, ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		rblob := accessio.BlobAccessForFile("text/plain", "/testdata/testcontent", env.FileSystem())
		dig := rblob.Digest()
		data, err = rblob.Get()
		Expect(err).To(Succeed())
		bpath := vfs.Join(nil, ARCH, comparch.BlobsDirectoryName, common.DigestToFileName(dig))
		Expect(vfs.FileExists(env.FileSystem(), bpath)).To(BeTrue())
		Expect(vfs.ReadFile(env.FileSystem(), bpath)).To(Equal(data))
		Expect(cd.Resources[0].Name).To(Equal("testdata"))
		Expect(cd.Resources[0].Version).To(Equal(VERSION))
		spec, err := env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)
		Expect(err).To(Succeed())
		Expect(spec.GetType()).To(Equal(accessmethods.LocalBlobType))
	})
})
