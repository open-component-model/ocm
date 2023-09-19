// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCHIVE = "archive"

var _ = Describe("blobhandler", func() {
	Context("regular", func() {
		var b *builder.Builder

		BeforeEach(func() {
			b = builder.NewBuilder()
		})

		AfterEach(func() {
			b.Cleanup()
		})

		It("uses generic local access", func() {
			b.ComponentArchive(ARCHIVE, accessio.FormatDirectory, "github.com/mandelsoft/test", "1.0.0", func() {
				b.Resource("test", "1.0.0", "Test", v1.LocalRelation, func() {
					b.BlobStringData(mime.MIME_TEXT, "testdata")
				})
			})
			data := Must(b.ReadFile(vfs.Join(b, ARCHIVE, compdesc.ComponentDescriptorFileName)))
			cd := Must(compdesc.Decode(data))
			Expect(cd.Resources[0].Access.GetType()).To(Equal(localblob.Type))

			data = Must(json.Marshal(cd.Resources[0].Access))
			found := Must(localblob.Decode(data))
			spec := localblob.New("sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50", "", mime.MIME_TEXT, nil)
			Expect(found).To(Equal(spec))
		})
	})
	Context("legacy", func() {
		var b *builder.Builder
		BeforeEach(func() {
			b = builder.NewBuilder(env.NewEnvironment())
			Expect(b.ConfigContext().GetAttributes().SetAttribute(compatattr.ATTR_KEY, true)).To(Succeed())
		})
		AfterEach(func() {
			b.Cleanup()
		})
		It("uses generic local access", func() {
			b.ComponentArchive(ARCHIVE, accessio.FormatDirectory, "github.com/mandelsoft/test", "1.0.0", func() {
				b.Resource("test", "1.0.0", "Test", v1.LocalRelation, func() {
					b.BlobStringData(mime.MIME_TEXT, "testdata")
				})
			})
			data := Must(b.ReadFile(vfs.Join(b, ARCHIVE, compdesc.ComponentDescriptorFileName)))
			cd := Must(compdesc.Decode(data))
			Expect(cd.Resources[0].Access.GetType()).To(Equal(localfsblob.Type))

			data = Must(json.Marshal(cd.Resources[0].Access))
			found := Must(localfsblob.Decode(data))

			spec := localfsblob.New("sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50", mime.MIME_TEXT)
			reflect.DeepEqual(found, spec)
			Expect(found).To(Equal(spec))
		})
	})

})
