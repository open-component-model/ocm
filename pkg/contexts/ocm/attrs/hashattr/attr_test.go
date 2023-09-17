// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hashattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/hashattr"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha512"
)

const NAME = "test"

var _ = Describe("attribute", func() {
	var cfgctx config.Context

	BeforeEach(func() {
		cfgctx = config.WithSharedAttributes(datacontext.New(nil)).New()
		_ = cfgctx
	})

	It("marshal/unmarshal", func() {
		cfg := hashattr.New(sha512.Algorithm)
		data := Must(json.Marshal(cfg))

		r := &hashattr.Config{}
		Expect(json.Unmarshal(data, r)).To(Succeed())
		Expect(r).To(Equal(cfg))
	})

	It("decode", func() {
		attr := &hashattr.Attribute{
			DefaultHasher: sha512.Algorithm,
		}

		r := Must(hashattr.AttributeType{}.Decode([]byte(sha512.Algorithm), runtime.DefaultYAMLEncoding))
		Expect(r).To(Equal(attr))
	})

	It("applies string", func() {
		MustBeSuccessful(cfgctx.GetAttributes().SetAttribute(hashattr.ATTR_KEY, sha512.Algorithm))
		attr := hashattr.Get(cfgctx)
		Expect(attr.GetHasher(cfgctx)).To(Equal(sha512.Handler{}))
	})

	It("applies config", func() {
		cfg := hashattr.New(sha512.Algorithm)

		MustBeSuccessful(cfgctx.ApplyConfig(cfg, "from test"))
		Expect(hashattr.Get(cfgctx).GetHasher(cfgctx)).To(Equal(sha512.Handler{}))
	})
})
