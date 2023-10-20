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
)

const COMPONENT = "acme.org/testcomp"
const VERSION = "1.0.0"

var _ = Describe("repository", func() {
	var ctx = ocm.DefaultContext()

	It("handles cvs", func() {
		finalize := finalizer.Finalizer{}
		defer Defer(finalize.Finalize)

		nested := finalize.Nested()

		repo := me.NewRepository(ctx)
		finalize.Close(repo, "source repo")

		c := Must(repo.LookupComponent(COMPONENT))
		finalize.Close(c, "src comp")

		cv := Must(c.NewVersion(VERSION))
		nested.Close(cv, "src vers")

		cv.GetDescriptor().Provider.Name = "acme.org"
		// wrap a non-closer access into a ref counting access to check cleanup
		blob := bpi.NewBlobAccessForBase(blobaccess.ForString(mime.MIME_TEXT, "testdata"))
		nested.Close(blob, "blob")
		MustBeSuccessful(cv.SetResourceBlob(ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blob, "", nil))
		MustBeSuccessful(c.AddVersion(cv))

		MustBeSuccessful(nested.Finalize())

		cv = Must(c.LookupVersion(VERSION))
		finalize.Close(cv, "query")
		rs := Must(cv.GetResourcesByName("test"))
		Expect(len(rs)).To(Equal(1))
		data := Must(ocmutils.GetResourceData(rs[0]))
		Expect(string(data)).To(Equal("testdata"))
	})
})
