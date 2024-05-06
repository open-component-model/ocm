package config_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("merge config", func() {

	spec1 := Must(v1.NewMergeAlgorithmSpecification("test", "config1"))
	spec2 := Must(v1.NewMergeAlgorithmSpecification("algo", "config2"))

	cfg := config.New()
	cfg.Assign("test", spec1)
	cfg.AssignLabel("l1", "v2", spec2)

	Context("serialize", func() {
		It("serializes config", func() {
			data := Must(json.Marshal(cfg))
			cfg2 := config.New()
			MustBeSuccessful(json.Unmarshal(data, cfg2))
			Expect(cfg2).To(Equal(cfg))
		})
	})

	Context("apply", func() {
		It("applies directly", func() {
			reg := hpi.NewRegistry()

			Expect(cfg.ApplyTo(nil, reg)).To(Succeed())

			found := reg.GetAssignments()
			expected := map[hpi.Hint]*hpi.Specification{
				"test":        spec1,
				"label:l1@v2": spec2,
			}

			Expect(found).To(DeepEqual(expected))
		})

		It("applies via config context", func() {
			ctx := ocm.New(datacontext.MODE_INITIAL)

			Expect(ctx.ConfigContext().ApplyConfig(cfg, "programmatic")).To(Succeed())

			Expect(valuemergehandler.For(ctx).GetHandlers()).To(Equal(valuemergehandler.Handlers{}))
			found := valuemergehandler.For(ctx).GetAssignments()
			expected := map[hpi.Hint]*hpi.Specification{
				"test":        spec1,
				"label:l1@v2": spec2,
			}
			Expect(found).To(DeepEqual(expected))
		})
	})
})
