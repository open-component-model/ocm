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
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociregistry"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart/loader"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch/comparch"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"

func CheckTextResource(env *TestEnv, cd *compdesc.ComponentDescriptor, name string) {
	rblob := accessio.BlobAccessForFile("text/plain", "/testdata/testcontent", env)
	dig := rblob.Digest()
	data, err := rblob.Get()
	Expect(err).To(Succeed())
	bpath := env.Join(ARCH, comparch.BlobsDirectoryName, common.DigestToFileName(dig))
	Expect(env.FileExists(bpath)).To(BeTrue())
	Expect(env.ReadFile(bpath)).To(Equal(data))

	r, err := cd.GetResourceByIdentity(metav1.NewIdentity(name))
	Expect(err).To(Succeed())
	Expect(r.Version).To(Equal(VERSION))
	Expect(r.Type).To(Equal("PlainText"))

	spec, err := env.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)
	Expect(err).To(Succeed())
	Expect(spec.GetType()).To(Equal(localblob.Type))
	Expect(spec.(*localblob.AccessSpec).LocalReference).To(Equal(common.DigestToFileName(dig)))
	Expect(spec.(*localblob.AccessSpec).MediaType).To(Equal("text/plain"))
}

func get(blob accessio.BlobAccess, expected []byte) []byte {
	data, err := blob.Get()
	ExpectWithOffset(1, err).To(Succeed())
	if expected != nil {
		ExpectWithOffset(1, string(data)).To(Equal(string(expected)))
	}
	return data
}

var _ = Describe("Add resources", func() {
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
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob by cli env file", func() {
		Expect(env.Execute("add", "resources", ARCH, "--settings", "/testdata/settings", "/testdata/resources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob by cli variable", func() {
		Expect(env.Execute("add", "resources", ARCH, "CONTENT=testcontent", "/testdata/resources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})
	It("adds helm chart", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/helm.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		r, err := cd.GetResourceByIdentity(metav1.NewIdentity("chart"))
		Expect(err).To(Succeed())
		Expect(r.Type).To(Equal("HelmChart"))
		Expect(r.Version).To(Equal(VERSION))

		Expect(r.Access.GetType()).To(Equal(localblob.Type))

		acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
		Expect(err).To(Succeed())
		// sha is always different for helm artefact
		// Expect(acc.(*localblob.AccessSpec).LocalReference).To(Equal("sha256.817db2696ed23f7779a7f848927e2958d2236e5483ad40875274462d8fa8ef9a"))

		blobpath := env.Join(ARCH, comparch.BlobsDirectoryName, common.DigestToFileName(digest.Digest(acc.(*localblob.AccessSpec).LocalReference)))
		blob := accessio.BlobAccessForFile(mime.MIME_GZIP, blobpath, env)

		set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
		Expect(err).To(Succeed())
		defer set.Close()
		art, err := set.GetArtefact(set.GetMain().String())
		Expect(err).To(Succeed())
		m := art.ManifestAccess().GetDescriptor()
		Expect(len(m.Layers)).To(Equal(1))

		blob, err = set.GetBlob(m.Layers[0].Digest)
		Expect(err).To(Succeed())
		reader, err := blob.Reader()
		Expect(err).To(Succeed())

		_, err = loader.LoadArchive(reader)
		Expect(err).To(Succeed())
	})

	It("adds external image", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/image.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		r, err := cd.GetResourceByIdentity(metav1.NewIdentity("image"))
		Expect(err).To(Succeed())
		Expect(r.Type).To(Equal("ociImage"))
		Expect(r.Version).To(Equal("v0.1.0"))
		Expect(r.Relation).To(Equal(metav1.ResourceRelation("external")))

		Expect(r.Access.GetType()).To(Equal(ociregistry.Type))

		acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociregistry.AccessSpec)(nil))))
		Expect(acc.(*ociregistry.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
	})
})
