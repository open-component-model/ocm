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

package genericocireg_test

import (
	"reflect"

	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = ocm.New()

const COMPONENT = "github.com/mandelsoft/ocm"

var _ = Describe("component repository mapping", func() {
	var tempfs vfs.FileSystem

	var ocispec oci.RepositorySpec
	var spec *genericocireg.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		ocispec = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessobj.PathFileSystem(tempfs), accessobj.FormatDirectory)
		spec = genericocireg.NewRepositorySpec(ocispec, nil)

	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("creates a dummy component", func() {
		repo, err := DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())

		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		Expect(repo.(*genericocireg.Repository).Close()).To(Succeed())

		// access it again
		repo, err = DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())

		ok, err := repo.ExistsComponentVersion(COMPONENT, "v1")
		Expect(err).To(Succeed())
		Expect(ok).To(BeTrue())

		comp, err = repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err = comp.LookupVersion("v1")
		Expect(err).To(Succeed())
		Expect(vers.GetDescriptor()).To(Equal(compdesc.New(COMPONENT, "v1")))

		Expect(repo.(*genericocireg.Repository).Close()).To(Succeed())
	})

})
