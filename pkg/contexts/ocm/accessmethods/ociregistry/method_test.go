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

package ociregistry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociregistry"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
)

const OCIPATH = "/tmp/oci"
const OCINAMESPACE = "ocm/test"
const OCIVERSION = "v2.0"
const OCIHOST = "alias"

type DummyAccess struct {
	ctx cpi.Context
}

var _ cpi.ComponentVersionAccess = (*DummyAccess)(nil)

func (d *DummyAccess) GetContext() core.Context {
	return d.ctx
}

func (d *DummyAccess) GetName() string {
	panic("implement me")
}

func (d *DummyAccess) GetVersion() string {
	panic("implement me")
}

func (d *DummyAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	panic("implement me")
}

func (d *DummyAccess) GetResources() []core.ResourceAccess {
	panic("implement me")
}

func (d *DummyAccess) GetResource(meta metav1.Identity) (core.ResourceAccess, error) {
	panic("implement me")
}

func (d *DummyAccess) GetSources() []core.SourceAccess {
	panic("implement me")
}

func (d *DummyAccess) GetSource(meta metav1.Identity) (core.SourceAccess, error) {
	panic("implement me")
}

func (d *DummyAccess) AccessMethod(spec core.AccessSpec) (core.AccessMethod, error) {
	panic("implement me")
}

func (d *DummyAccess) AddBlob(blob core.BlobAccess, refName string, global core.AccessSpec) (core.AccessSpec, error) {
	panic("implement me")
}

func (d *DummyAccess) SetResourceBlob(meta *core.ResourceMeta, blob core.BlobAccess, refname string, global core.AccessSpec) error {
	panic("implement me")
}

func (d *DummyAccess) SetResource(meta *core.ResourceMeta, spec compdesc.AccessSpec) error {
	panic("implement me")
}

func (d *DummyAccess) SetSourceBlob(meta *core.SourceMeta, blob core.BlobAccess, refname string, global core.AccessSpec) error {
	panic("implement me")
}

func (d *DummyAccess) SetSource(meta *core.SourceMeta, spec compdesc.AccessSpec) error {
	panic("implement me")
}

func (d *DummyAccess) SetReference(ref *core.ComponentReference) error {
	panic("implement me")
}

func (d *DummyAccess) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("Method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses artefact", func() {
		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			env.Namespace(OCINAMESPACE, func() {
				env.Manifest(OCIVERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "manifestlayer")
					})
				})
			})
		})

		env.OCIContext().SetAlias(OCIHOST, ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCIPATH, accessio.PathFileSystem(env.FileSystem())))

		spec := ociregistry.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION))

		m, err := spec.AccessMethod(&DummyAccess{env.OCMContext()})
		Expect(err).To(Succeed())

		Expect(m.(accessio.DigestSource).Digest().String()).To(Equal("sha256:0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f"))
	})
})
