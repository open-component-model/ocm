package config_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/download"
	me "ocm.software/ocm/api/ocm/extensions/download/config"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/tech/helm"
)

var _ = Describe("Download Handler regigistration", func() {
	It("register by ocm config", func() {
		ctx := ocm.New()

		cfg := me.New()
		cfg.AddRegistration(me.Registration{
			Name:        "helm/artifact",
			Description: "some registration",
			HandlerOptions: download.HandlerOptions{
				HandlerKey: download.HandlerKey{
					ArtifactType: "someType",
				},
			},
			Config: nil,
		})

		data := Must(json.Marshal(cfg))
		ocmutils.ConfigureByData(ctx, data, "manual")

		h := download.For(ctx).LookupHandler("someType", helm.ChartMediaType)
		Expect(h.Len()).To(Equal(1))
	})
})
