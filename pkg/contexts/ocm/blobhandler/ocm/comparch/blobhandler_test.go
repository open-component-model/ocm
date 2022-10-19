// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"
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
	"github.com/open-component-model/ocm/pkg/runtime"
)

const ARCHIVE = "archive"

var _ = Describe("blobhandler", func() {
	log := logging.DefaultContext().Logger()

	Context("regular", func() {
		var b *builder.Builder

		BeforeEach(func() {
			b = builder.NewBuilder(env.NewEnvironment())
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
			data, err := b.ReadFile(vfs.Join(b, ARCHIVE, compdesc.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(cd.Resources[0].Access.GetType()).To(Equal(localblob.Type))

			data, err = json.Marshal(cd.Resources[0].Access)
			Expect(err).To(Succeed())
			log.Info("marshal result", "data", string(data))
			found := &localblob.AccessSpec{}
			Expect(json.Unmarshal(data, found)).To(Succeed())

			spec := &localblob.AccessSpec{
				ObjectVersionedType: runtime.NewVersionedObjectType(localblob.Type),
				MediaType:           mime.MIME_TEXT,
				LocalReference:      "sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50",
			}
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
			data, err := b.ReadFile(vfs.Join(b, ARCHIVE, compdesc.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(cd.Resources[0].Access.GetType()).To(Equal(localfsblob.Type))

			data, err = json.Marshal(cd.Resources[0].Access)
			Expect(err).To(Succeed())
			log.Info("marshal result", "data", string(data))
			found := &localfsblob.AccessSpec{}
			Expect(json.Unmarshal(data, found)).To(Succeed())

			spec := &localfsblob.AccessSpec{
				ObjectVersionedType: runtime.NewVersionedObjectType(localfsblob.Type),
				MediaType:           mime.MIME_TEXT,
				Filename:            "sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50",
			}
			reflect.DeepEqual(found, spec)
			Expect(found).To(Equal(spec))
		})
	})

})
