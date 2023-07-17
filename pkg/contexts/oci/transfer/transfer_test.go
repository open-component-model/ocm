// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer"
	"github.com/open-component-model/ocm/pkg/finalizer"
)

const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"

var _ = Describe("transfer OCI artifacts", func() {

	var env *Builder
	var idesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			idesc = OCIIndex1(env)
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers index", func() {
		// index implicitly tests transfer of simple manifest, also
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		src := Must(ctf.Open(env.OCIContext(), accessobj.ACC_READONLY, OCIPATH, 0, env))
		finalize.Close(src, "source")
		art := Must(src.LookupArtifact(OCINAMESPACE3, OCIINDEXVERSION))
		finalize.Close(art, "source artifact")

		tgt := Must(ctf.Create(env.OCIContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer Close(tgt, "target")
		ns := Must(tgt.LookupNamespace(OCINAMESPACE3))
		defer Close(ns, "target namespace")

		MustBeSuccessful(transfer.TransferArtifact(art, ns, OCIINDEXVERSION))

		MustBeSuccessful(finalize.Finalize())

		tart := Must(ns.GetArtifact(idesc.Digest.String()))
		defer Close(tart, "target index artifact")

		Expect(tart.IsIndex())
		manifests := tart.IndexAccess().GetDescriptor().Manifests
		Expect(len(manifests)).To(Equal(2))
		Expect(manifests[0].Digest.Encoded()).To(Equal(D_OCIMANIFEST1))
		Expect(manifests[1].Digest.Encoded()).To(Equal(D_OCIMANIFEST2))

		nart1 := Must(ns.GetArtifact(manifests[0].Digest.String()))
		defer Close(nart1, "nested artifact 1")
		Expect(nart1.IsManifest()).To(BeTrue())
		Expect(len(nart1.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
		blob := Must(nart1.ManifestAccess().GetBlob(nart1.ManifestAccess().GetDescriptor().Layers[0].Digest))
		defer Close(blob, "layer 0 of nested artifact 1")
		data := Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER))

		nart2 := Must(ns.GetArtifact(manifests[1].Digest.String()))
		defer Close(nart2, "nested artifact 2")
		Expect(nart2.IsManifest()).To(BeTrue())
		Expect(len(nart2.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
		blob = Must(nart2.ManifestAccess().GetBlob(nart2.ManifestAccess().GetDescriptor().Layers[0].Digest))
		defer Close(blob, "layer 0 of nested artifact 2")
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER2))
	})
})
