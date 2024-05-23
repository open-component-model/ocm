// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helmblob_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/helmblob"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("", func() {
	var e *Builder

	BeforeEach(func() {
		e = NewBuilder(env.TestData())

	})

	AfterEach(func() {
		MustBeSuccessful(e.Cleanup())
	})

	It("", func() {
		Expect(1).To(Equal(1))
		ctf := Must(ctfocm.Open(e, accessobj.ACC_CREATE, "/repo", 0o700, e, ctfocm.FormatDirectory))
		defer Close(ctf)
		cv := Must(ctf.NewComponentVersion("ocm.software/test-component", "1.0.0"))
		defer Close(cv)
		MustBeSuccessful(cv.SetResourceAccess(me.ResourceAccess(e.OCMContext(), cpi.NewResourceMeta("helm1", "blob", metav1.LocalRelation), "/testdata/testchart1", me.WithFileSystem(e.FileSystem()))))
		MustBeSuccessful(cv.SetResourceAccess(me.ResourceAccess(e.OCMContext(), cpi.NewResourceMeta("helm2", "blob", metav1.LocalRelation), "/testdata/testchart2", me.WithFileSystem(e.FileSystem()))))
		MustBeSuccessful(ctf.AddComponentVersion(cv))
	})
})
