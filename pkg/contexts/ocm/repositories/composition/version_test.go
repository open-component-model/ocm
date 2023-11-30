// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package composition_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

var _ = Describe("version", func() {
	var ctx = ocm.DefaultContext()

	It("handles anonymous version", func() {
		finalize := finalizer.Finalizer{}
		defer Defer(finalize.Finalize)

		nested := finalize.Nested()

		// compose new version
		cv := me.NewComponentVersion(ctx, COMPONENT, VERSION)
		cv.GetDescriptor().Provider.Name = "acme.org"
		finalize.Close(cv, "composed version")

		// wrap a non-closer access into a ref counting access to check cleanup
		blob := bpi.NewBlobAccessForBase(blobaccess.ForString(mime.MIME_TEXT, "testdata"))
		nested.Close(blob, "blob")
		MustBeSuccessful(cv.SetResourceBlob(ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blob, "", nil))

		// add version to repository
		repo1 := me.NewRepository(ctx)
		finalize.Close(repo1, "target repo1")
		c := Must(repo1.LookupComponent(COMPONENT))
		finalize.Close(c, "src comp")
		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(nested.Finalize())

		// check result
		cv = Must(c.LookupVersion(VERSION))
		Expect(refmgmt.ReferenceCount(cv)).To(Equal(1))
		nested.Close(cv, "query")
		rs := Must(cv.GetResourcesByName("test"))
		Expect(len(rs)).To(Equal(1))
		data := Must(ocmutils.GetResourceData(rs[0]))
		Expect(string(data)).To(Equal("testdata"))
		Expect(refmgmt.ReferenceCount(cv)).To(Equal(1))

		// add this version again
		repo2 := me.NewRepository(ctx)
		finalize.Close(repo2, "target repo2")
		MustBeSuccessful(repo2.AddComponentVersion(cv))
		Expect(refmgmt.ReferenceCount(cv)).To(Equal(1))
		MustBeSuccessful(nested.Finalize())

		// check result
		cv = Must(repo2.LookupComponentVersion(COMPONENT, VERSION))
		finalize.Close(cv, "query")
		rs = Must(cv.GetResourcesByName("test"))
		Expect(len(rs)).To(Equal(1))
		data = Must(ocmutils.GetResourceData(rs[0]))
		Expect(string(data)).To(Equal("testdata"))
	})
})
