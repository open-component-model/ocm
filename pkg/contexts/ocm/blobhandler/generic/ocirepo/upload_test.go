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

package ocirepo_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	tenv "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
)

const COMP = "github.com/compa"
const VERS = "1.0.0"
const CA = "ca"
const CTF = "ctf"
const COPY = "ctf.copy"
const TARGET = "/tmp/target"

func Close(closer io.Closer) {
	err := closer.Close()
	ExpectWithOffset(1, err).To(Succeed())
}

const OCIHOST = "alias"
const OCIPATH = "/tmp/source"
const OCINAMESPACE = "ocm/value"
const OCIVERSION = "v2.0"

var _ = Describe("upload", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(tenv.NewEnvironment())

		// fake OCI registry
		env.OCIContext().SetAlias(OCIHOST, ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCIPATH, accessio.PathFileSystem(env.FileSystem())))
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

		env.OCICommonTransport(TARGET, accessio.FormatDirectory)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers oci artefact", func() {
		ctx := env.OCMContext()

		env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERS, func() {
			env.Provider("mandelsoft")
			env.Resource("value", "", resourcetypes.OCI_IMAGE, v1.LocalRelation, func() {
				env.Access(
					ociartefact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
				)
			})
		})
		ca, err := comparch.Open(ctx, accessobj.ACC_READONLY, CA, 0, env)
		Expect(err).To(Succeed())
		oca := accessio.OnceCloser(ca)
		defer Close(oca)

		ctf, err := ctfocm.Create(ctx, accessobj.ACC_CREATE, CTF, 0700, env)
		Expect(err).To(Succeed())
		octf := accessio.OnceCloser(ctf)
		defer Close(octf)

		handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())

		err = transfer.TransferVersion(nil, nil, ca.Repository(), ca, ctf, handler)
		Expect(err).To(Succeed())
		oca.Close()

		// now we have a transport archive with local blob for the image

		cv, err := ctf.LookupComponentVersion(COMP, VERS)
		Expect(err).To(Succeed())
		ocv := accessio.OnceCloser(cv)
		defer Close(ocv)
		ra, err := cv.GetResourceByIndex(0)
		Expect(err).To(Succeed())
		acc, err := ra.Access()
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		// transfer component
		copy, err := ctfocm.Create(ctx, accessobj.ACC_CREATE, COPY, 0700, env)
		Expect(err).To(Succeed())
		ocopy := accessio.OnceCloser(copy)
		defer Close(ocopy)

		// prepare upload to target OCI repo
		attr := ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")
		ociuploadattr.Set(ctx, attr)

		err = transfer.TransferVersion(nil, nil, ctf, cv, copy, nil)
		Expect(err).To(Succeed())

		// check type
		cv2, err := copy.LookupComponentVersion(COMP, VERS)
		Expect(err).To(Succeed())
		ocv2 := accessio.OnceCloser(cv2)
		defer Close(ocv2)
		ra, err = cv2.GetResourceByIndex(0)
		Expect(err).To(Succeed())
		acc, err = ra.Access()
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(ociartefact.Type))
		val, err := ctx.AccessSpecForSpec(acc)
		Expect(err).To(Succeed())
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartefact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0"))

		attr.Close()
		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtefact("copy/ocm/value", "v2.0")).To(BeTrue())

	})
})
