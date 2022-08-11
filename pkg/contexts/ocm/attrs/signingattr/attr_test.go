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

package signingattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
)

const NAME = "test"

var _ = Describe("attribute", func() {
	var cfgctx config.Context

	BeforeEach(func() {
		cfgctx = config.WithSharedAttributes(datacontext.New(nil)).New()
		_ = cfgctx
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
		Expect(signingattr.Get(cfgctx).GetPublicKey(NAME)).To(Equal([]byte("keydata")))
	})

})
