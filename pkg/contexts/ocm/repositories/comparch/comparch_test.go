// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch_test

import (
	"encoding/json"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const (
	TEST_FILEPATH     = "testfilepath"
	TAR_COMPARCH      = "testdata/common"
	DIR_COMPARCH      = "testdata/directory"
	RESOURCE_NAME     = "test"
	COMPONENT_NAME    = "example.com/root"
	COMPONENT_VERSION = "1.0.0"
)

var _ = Describe("Repository", func() {

	It("marshal/unmarshal simple", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TEST_FILEPATH))
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"ComponentArchive\",\"filePath\":\"testfilepath\",\"accessMode\":1}"))
		_ = Must(octx.RepositorySpecForConfig(data, runtime.DefaultJSONEncoding)).(*comparch.RepositorySpec)
		// spec will not equal r as the filesystem cannot be serialized
	})

	It("component archive with resource stored as tar", func() {
		// this is the typical use case
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TAR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		defer Close(cv)
		res := Must(cv.GetResourcesByName(RESOURCE_NAME))
		acc := Must(res[0].AccessMethod())
		defer Close(acc)
		bytesA := Must(acc.Get())

		bytesB := Must(vfs.ReadFile(osfs.New(), filepath.Join(TAR_COMPARCH, "blobs", "sha256.3ed99e50092c619823e2c07941c175ea2452f1455f570c55510586b387ec2ff2")))
		Expect(bytesA).To(Equal(bytesB))
	})

	It("component archive with a resource stored in a directory", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, DIR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		defer Close(cv)
		res := Must(cv.GetResourcesByName(RESOURCE_NAME))
		acc := Must(res[0].AccessMethod())
		defer Close(acc)
		data := Must(acc.Reader())
		defer Close(data)

		mfs := memoryfs.New()
		_, _, err := tarutils.ExtractTarToFsWithInfo(mfs, data)
		Expect(err).ToNot(HaveOccurred())
		bufferA := Must(vfs.ReadFile(mfs, "testfile"))
		bufferB := Must(vfs.ReadFile(osfs.New(), filepath.Join(DIR_COMPARCH, "blobs", "root", "testfile")))
		Expect(bufferA).To(Equal(bufferB))
	})

	It("closing a resource before actually reading it", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TAR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		defer Close(cv)
		res := Must(cv.GetResourcesByName(RESOURCE_NAME))
		acc := Must(res[0].AccessMethod())
		defer Close(acc)
	})
})
