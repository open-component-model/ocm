// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/clictx/config"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	ocmocireg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
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

	ocmspec := ocmocireg.NewRepositorySpec("gcr.io", nil)
	ocmdata, err := normalize(ocmspec)
	Expect(err).To(Succeed())

	specdata := "{\"ociRepositories\":{\"oci\":" + string(ocidata) + "},\"ocmRepositories\":{\"ocm\":" + string(ocmdata) + "},\"type\":\"" + config.OCMCmdConfigType + "\"}"

	Context("serialize", func() {

		It("serializes config", func() {

			cfg := config.New()
			err := cfg.AddOCIRepository("oci", ocispec)
			Expect(err).To(Succeed())
			err = cfg.AddOCMRepository("ocm", ocmspec)
			Expect(err).To(Succeed())

			data, err := normalize(cfg)

			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(specdata)))

			cfg2 := config.New()
			err = json.Unmarshal(data, cfg2)
			Expect(err).To(Succeed())
			Expect(cfg2).To(Equal(cfg))
		})
	})
})
