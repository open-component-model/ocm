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

package config_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
)

func normalize(i interface{}) ([]byte, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	var generic map[string]interface{}
	err = json.Unmarshal(data, &generic)
	if err != nil {
		return nil, err
	}
	return json.Marshal(generic)
}

var _ = Describe("oci config", func() {

	spec := ocireg.NewRepositorySpec("gcr.io", nil)
	data, err := normalize(spec)
	Expect(err).To(Succeed())

	specdata := "{\"aliases\":{\"alias\":" + string(data) + "},\"type\":\"" + config.ConfigType + "\"}"

	Context("serialize", func() {
		It("serializes config", func() {
			cfg := config.NewConfigSpec()
			err := cfg.SetAlias("alias", spec)
			Expect(err).To(Succeed())

			data, err := normalize(cfg)

			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(specdata)))

			cfg2 := config.NewConfigSpec()
			err = json.Unmarshal(data, cfg2)
			Expect(err).To(Succeed())
			Expect(cfg2).To(Equal(cfg))
		})
	})

	Context("apply", func() {
		It("applies directly", func() {
			ctx := cpi.New()

			cfg := config.NewConfigSpec()
			err := cfg.SetAlias("alias", spec)
			Expect(err).To(Succeed())

			Expect(cfg.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())

			found := ctx.GetAlias("alias")
			Expect(found).To(Equal(cfg.Aliases["alias"]))
		})

		It("applies via config context", func() {
			ctx := cpi.New()

			cfg := config.NewConfigSpec()
			err := cfg.SetAlias("alias", spec)
			Expect(err).To(Succeed())

			Expect(ctx.ConfigContext().ApplyConfig(cfg, "programmatic")).To(Succeed())

			found := ctx.GetAlias("alias")
			Expect(found).To(Equal(cfg.Aliases["alias"]))
		})
	})
})
