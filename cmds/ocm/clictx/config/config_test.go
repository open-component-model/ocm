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

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/clictx/config"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	ocmocireg "github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = clictx.New()

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

var _ = Describe("command config", func() {

	ocispec := ocireg.NewRepositorySpec("ghcr.io")

	ocidata, err := normalize(ocispec)
	Expect(err).To(Succeed())

	ocmspec := ocmocireg.NewRepositorySpec("gcr.io", "")
	ocmdata, err := normalize(ocmspec)
	Expect(err).To(Succeed())

	specdata := "{\"ociRepositories\":{\"oci\":" + string(ocidata) + "},\"ocmRepositories\":{\"ocm\":" + string(ocmdata) + "},\"type\":\"" + config.OCMCmdConfigType + "\"}"

	Context("serialize", func() {

		It("serializes rconfig", func() {

			cfg := config.NewConfigSpec()
			err := cfg.AddOCIRepository("oci", ocispec)
			Expect(err).To(Succeed())
			err = cfg.AddOCMRepository("ocm", ocmspec)
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
})
