// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/utils"
)

const COMPONENT = "github.com/mandelsoft/ocm"
const VERSION = "1.0.0"

var _ = Describe("access method", func() {
	var fs vfs.FileSystem
	ctx := ocm.DefaultContext()

	BeforeEach(func() {
		fs = memoryfs.New()
	})

	It("adds component  version", func() {
		final := utils.Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		a = Must(ctf.Open(ctx, accessobj.ACC_READONLY, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)

		cv = Must(a.LookupComponentVersion(COMPONENT, VERSION))
		final.Close(cv)
	})

	It("adds omits unadded new component version", func() {
		final := utils.Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		MustBeSuccessful(final.Finalize())

		a = Must(ctf.Open(ctx, accessobj.ACC_READONLY, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)

		_, err := a.LookupComponentVersion(COMPONENT, VERSION)

		Expect(err).To(MatchError(ContainSubstring("component version \"github.com/mandelsoft/ocm:1.0.0\" not found: oci artifact \"1.0.0\" not found in component-descriptors/github.com/mandelsoft/ocm")))
	})

})
