package config_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm/config"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
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
			cfg := config.New()
			err := cfg.SetAlias("alias", spec)
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

	Context("apply", func() {
		It("applies directly", func() {
			ctx := cpi.New()

			cfg := config.New()
			err := cfg.SetAlias("alias", spec)
			Expect(err).To(Succeed())

			Expect(cfg.ApplyTo(ctx.ConfigContext(), ctx)).To(Succeed())

			found := ctx.GetAlias("alias")
			Expect(found).To(Equal(cfg.Aliases["alias"]))
		})

		It("applies via config context", func() {
			ctx := cpi.New()

			cfg := config.New()
			err := cfg.SetAlias("alias", spec)
			Expect(err).To(Succeed())

			Expect(ctx.ConfigContext().ApplyConfig(cfg, "programmatic")).To(Succeed())

			found := ctx.GetAlias("alias")
			Expect(found).To(Equal(cfg.Aliases["alias"]))
		})
	})
})
