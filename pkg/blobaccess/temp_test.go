// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blobaccess_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
)

var _ = Describe("temp file management", func() {
	var tempfs vfs.FileSystem

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("temp file exists", func() {
		tmp, err := blobaccess.NewTempFile("", "test*.tmp", tempfs)
		Expect(err).To(Succeed())
		defer tmp.Close()
		Expect(vfs.FileExists(tempfs, tmp.Name())).To(BeTrue())
	})
	It("temp file deleted on close", func() {
		tmp, err := blobaccess.NewTempFile("", "test*.tmp", tempfs)
		Expect(err).To(Succeed())
		name := tmp.Name()
		Expect(tmp.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeFalse())
	})
	It("temp file released", func() {
		tmp, err := blobaccess.NewTempFile("", "test*.tmp", tempfs)
		Expect(err).To(Succeed())
		name := tmp.Name()
		file := tmp.Release()
		defer file.Close()
		Expect(tmp.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeTrue())
	})
	It("temp file blob", func() {
		tmp, err := blobaccess.NewTempFile("", "test*.tmp", tempfs)
		Expect(err).To(Succeed())
		name := tmp.Name()
		blob := tmp.AsBlob("ttt")
		Expect(tmp.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeTrue())
		Expect(blob.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeFalse())
	})
	It("temp file blob access", func() {
		value := []byte("this is a test")
		tmp, err := blobaccess.NewTempFile("", "test*.tmp", tempfs)
		Expect(err).To(Succeed())
		tmp.Writer().Write(value)
		Expect(tmp.Sync()).To(Succeed())
		name := tmp.Name()
		blob := tmp.AsBlob("ttt")
		Expect(tmp.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeTrue())
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect(data).To(Equal(value))
		data, err = blob.Get()
		Expect(err).To(Succeed())
		Expect(data).To(Equal(value))
		Expect(blob.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeFalse())
	})
})
