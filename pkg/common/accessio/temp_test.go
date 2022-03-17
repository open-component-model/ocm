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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("artefact management", func() {
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
		tmp, err := accessio.NewTempFile(tempfs, "", "test*.tmp")
		Expect(err).To(Succeed())
		defer tmp.Close()
		Expect(vfs.FileExists(tempfs, tmp.Name())).To(BeTrue())
	})
	It("temp file deleted on close", func() {
		tmp, err := accessio.NewTempFile(tempfs, "", "test*.tmp")
		Expect(err).To(Succeed())
		name := tmp.Name()
		Expect(tmp.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeFalse())
	})
	It("temp file released", func() {
		tmp, err := accessio.NewTempFile(tempfs, "", "test*.tmp")
		Expect(err).To(Succeed())
		name := tmp.Name()
		file := tmp.Release()
		defer file.Close()
		Expect(tmp.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, name)).To(BeTrue())
	})
	It("temp file blob", func() {
		tmp, err := accessio.NewTempFile(tempfs, "", "test*.tmp")
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
		tmp, err := accessio.NewTempFile(tempfs, "", "test*.tmp")
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
