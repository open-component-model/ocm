// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch_test

import (
	"bytes"
	"encoding/json"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/runtime"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const (
	TEST_FILEPATH = "testfilepath"
	TAR_COMPARCH  = "testdata/common"
	DIR_COMPARCH  = "testdata/directory"
)

var _ = Describe("Repository", func() {

	It("marshal/unmarshal simple", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_WRITABLE, TEST_FILEPATH))
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"" + comparch.Type + "\",\"filePath\":\"" + TEST_FILEPATH + "\"}"))
		_ = Must(octx.RepositorySpecForConfig(data, runtime.DefaultJSONEncoding)).(*comparch.RepositorySpec)
		// spec will not equal r as the filesystem cannot be serialized
	})

	It("component archive with resource stored as tar", func() {
		// this is the typical use case
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_WRITABLE, TAR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		cv := Must(repo.LookupComponentVersion("example.com/root", "1.0.0"))
		res := Must(cv.GetResourcesByName("root-a"))
		acc := Must(res[0].AccessMethod())
		defer acc.Close()
		data := Must(acc.Reader())
		defer data.Close()

		mfs := memoryfs.New()
		data, _ = Must2(compression.AutoDecompress(data))
		_, _ = Must2(tarutils.ExtractTarToFsWithInfo(mfs, data))
		bytesA := []byte{}
		_ = Must(Must(mfs.Open("blueprint.yaml")).Read(bytesA))

		bytesB := []byte{}
		_ = Must(Must(osfs.New().Open(TAR_COMPARCH + "/blobs/sha256.aeb2b713150dc9baa889184a406297990259f3919d7dd644cbfe49cd352a2a44")).Read(bytesB))
		bufferB := bytes.NewBuffer(bytesB)
		r, _ := Must2(compression.AutoDecompress(bufferB))
		_, _ = Must2(tarutils.ExtractTarToFsWithInfo(mfs, r))
		Expect(bytesA).To(Equal(bufferB.Bytes()))
	})

	It("component archive with a resource stored in a directory", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_WRITABLE, DIR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		cv := Must(repo.LookupComponentVersion("example.com/root", "1.0.0"))
		res := Must(cv.GetResourcesByName("root-a"))
		acc := Must(res[0].AccessMethod())
		defer acc.Close()
		data := Must(acc.Reader())
		defer data.Close()

		mfs := memoryfs.New()
		_, _, err := tarutils.ExtractTarToFsWithInfo(mfs, data)
		Expect(err).ToNot(HaveOccurred())
		bufferA := []byte{}
		bufferB := []byte{}
		_ = Must(Must(mfs.Open("blueprint.yaml")).Read(bufferA))
		_ = Must(Must(osfs.New().Open(DIR_COMPARCH + "/blobs/root/blueprint.yaml")).Read(bufferB))
		Expect(bufferA).To(Equal(bufferB))
	})
})
