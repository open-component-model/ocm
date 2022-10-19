// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cacheattr_test

import (
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("attribute", func() {
	var ctx ocm.Context
	var cfgctx config.Context
	var cache accessio.BlobCache

	BeforeEach(func() {
		var err error
		cfgctx = config.WithSharedAttributes(datacontext.New(nil)).New()
		credctx := credentials.WithConfigs(cfgctx).New()
		ocictx := oci.WithCredentials(credctx).New()
		ctx = ocm.WithOCIRepositories(ocictx).New()
		cache, err = accessio.NewDefaultBlobCache()
		Expect(err).To(Succeed())
	})
	AfterEach(func() {
		cache.Unref()
	})

	It("local setting", func() {
		Expect(cacheattr.Get(ctx)).To(BeNil())
		Expect(cacheattr.Set(ctx, cache)).To(Succeed())
		Expect(cacheattr.Get(ctx)).To(BeIdenticalTo(cache))
	})

	It("global setting", func() {
		Expect(cacheattr.Get(cfgctx)).To(BeNil())
		Expect(cacheattr.Set(ctx, cache)).To(Succeed())
		Expect(cacheattr.Get(ctx)).To(BeIdenticalTo(cache))
	})

	It("parses string", func() {
		dir := os.TempDir()
		cache, err := cacheattr.AttributeType{}.Decode([]byte(dir), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(cache).String()).To(Equal("*accessio.blobCache"))
	})
})
