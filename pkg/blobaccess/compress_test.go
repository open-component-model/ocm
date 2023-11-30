// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blobaccess_test

import (
	"bytes"
	"compress/gzip"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("temp file management", func() {

	Context("compress", func() {
		It("compress access", func() {
			blob := blobaccess.ForString(mime.MIME_TEXT, "testdata")
			defer blob.Close()

			comp := Must(blobaccess.WithCompression(blob))
			defer comp.Close()

			Expect(comp.MimeType()).To(Equal(mime.MIME_TEXT + "+gzip"))
			data := Must(comp.Get())
			Expect(len(data)).To(Not(Equal(8)))

			uncomp := Must(io.ReadAll(Must(gzip.NewReader(bytes.NewReader(data)))))
			Expect(string(uncomp)).To(Equal("testdata"))
		})

		It("compress reader access", func() {
			blob := blobaccess.ForString(mime.MIME_TEXT, "testdata")
			defer blob.Close()

			comp := Must(blobaccess.WithCompression(blob))
			defer comp.Close()

			r := Must(comp.Reader())
			data := Must(io.ReadAll(r))
			Expect(len(data)).To(Not(Equal(8)))

			uncomp := Must(io.ReadAll(Must(gzip.NewReader(bytes.NewReader(data)))))
			Expect(string(uncomp)).To(Equal("testdata"))
		})
	})

	Context("uncompress", func() {
		buf := bytes.NewBuffer(nil)
		cw := gzip.NewWriter(buf)
		MustBeSuccessful(io.WriteString(cw, "testdata"))
		cw.Close()

		It("uncompress access", func() {
			blob := blobaccess.ForData(mime.MIME_TEXT+"+gzip", buf.Bytes())
			defer blob.Close()

			comp := Must(blobaccess.WithDecompression(blob))
			defer comp.Close()
			Expect(comp.MimeType()).To(Equal(mime.MIME_TEXT))

			data := Must(comp.Get())
			Expect(string(data)).To(Equal("testdata"))
		})

		It("compress reader access", func() {
			blob := blobaccess.ForData(mime.MIME_TEXT+"+gzip", buf.Bytes())
			defer blob.Close()

			comp := Must(blobaccess.WithDecompression(blob))
			defer comp.Close()

			r := Must(comp.Reader())
			data := Must(io.ReadAll(r))
			Expect(string(data)).To(Equal("testdata"))
		})
	})

})
