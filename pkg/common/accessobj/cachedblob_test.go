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

package accessobj_test

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/tmpcache"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/mime"
)

type Source struct {
	accessio.DataSource
	count int
}

func NewData(data string) *Source {
	return &Source{accessio.DataAccessForString(data), 0}
}

func (s *Source) Reader() (io.ReadCloser, error) {
	s.count++
	return s.DataSource.Reader()
}

func (s *Source) Count() int {
	return s.count
}

type WriteAtSource struct {
	accessio.DataWriter
	data  []byte
	count int
}

func NewWriteAtSource(data string) *WriteAtSource {
	w := &WriteAtSource{data: []byte(data), count: 0}
	w.DataWriter = accessio.NewWriteAtWriter(w.write)
	return w
}

func (s *WriteAtSource) write(writer io.WriterAt) error {
	s.count++
	_, err := writer.WriteAt(s.data, 0)
	return err
}

func (s *WriteAtSource) Count() int {
	return s.count
}

var _ = Describe("cached blob", func() {
	var fs vfs.FileSystem
	var ctx datacontext.Context

	BeforeEach(func() {
		var err error
		fs, err = osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		ctx = datacontext.New(nil)
		vfsattr.Set(ctx, fs)
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: "/tmp"})
	})

	AfterEach(func() {
		vfs.Cleanup(fs)
	})

	It("reads data source only once", func() {
		src := NewData("testdata")
		blob := accessobj.CachedBlobAccessForDataAccess(ctx, mime.MIME_TEXT, src)

		Expect(blob.Size()).To(Equal(int64(8)))
		Expect(blob.Digest().String()).To(Equal("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(src.Count()).To(Equal(1))
		dir, err := vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(1))
		Expect(blob.Close()).To(Succeed())

		dir, err = vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(0))
	})

	It("reads write at source only once", func() {
		src := NewWriteAtSource("testdata")
		blob := accessobj.CachedBlobAccessForWriter(ctx, mime.MIME_TEXT, src)

		Expect(blob.Size()).To(Equal(int64(8)))
		Expect(blob.Digest().String()).To(Equal("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(blob.Get()).To(Equal([]byte("testdata")))
		Expect(src.Count()).To(Equal(1))
		dir, err := vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(1))
		Expect(blob.Close()).To(Succeed())

		dir, err = vfs.ReadDir(fs, "/tmp")
		Expect(err).To(Succeed())
		Expect(len(dir)).To(Equal(0))
	})

})
