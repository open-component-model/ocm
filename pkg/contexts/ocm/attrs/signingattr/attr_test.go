// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signingattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
)

const NAME = "test"

var _ = Describe("attribute", func() {
	var cfgctx config.Context
	var ocmctx ocm.Context

	BeforeEach(func() {
		ocmctx = ocm.New(datacontext.MODE_EXTENDED)
		cfgctx = ocmctx.ConfigContext()
	})

	It("marshal/unmarshal", func() {
		cfg := signingattr.New()
		cfg.AddPublicKeyData(NAME, []byte("keydata"))

		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		r := &signingattr.Config{}
		Expect(json.Unmarshal(data, r)).To(Succeed())
		Expect(r).To(Equal(cfg))
	})

	It("applies", func() {
		cfg := signingattr.New()
		cfg.AddPublicKeyData(NAME, []byte("keydata"))

		Expect(cfgctx.ApplyConfig(cfg, "from test")).To(Succeed())
		Expect(signingattr.Get(ocmctx).GetPublicKey(NAME)).To(Equal([]byte("keydata")))
	})

})
