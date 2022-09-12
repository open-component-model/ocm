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

package accessio_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
)

var _ = Describe("cache management", func() {
	var tempfs vfs.FileSystem
	var cache accessio.BlobCache
	var source accessio.BlobCache

	var td1_digest digest.Digest
	var td1_size int64

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t
		local, err := accessio.NewDefaultBlobCache(t)
		Expect(err).To(Succeed())

		source, err = accessio.NewDefaultBlobCache()
		Expect(err).To(Succeed())

		td1_size, td1_digest, err = source.AddData(accessio.DataAccessForBytes([]byte("testdata")))
		Expect(err).To(Succeed())

		cache, err = accessio.CachedAccess(source, nil, local)
		Expect(err).To(Succeed())

		_ = td1_size
	})

	AfterEach(func() {
		cache.Unref()
		source.Unref()
		vfs.Cleanup(tempfs)
	})

	It("blob copied to cache", func() {
		Expect(vfs.FileExists(tempfs, common.DigestToFileName(td1_digest))).To(BeFalse())
		_, data, err := cache.GetBlobData(td1_digest)
		Expect(err).To(Succeed())
		Expect(vfs.FileExists(tempfs, common.DigestToFileName(td1_digest))).To(BeFalse())
		Expect(data.Get()).To(Equal([]byte("testdata")))
		Expect(vfs.FileExists(tempfs, common.DigestToFileName(td1_digest))).To(BeTrue())
	})
})
