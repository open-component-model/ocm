package helm_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/helm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	helm2 "github.com/open-component-model/ocm/pkg/helm"
)

var _ = Describe("Method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses artifact", func() {
		resp, err := http.Get("https://charts.helm.sh/stable")
		if err == nil { // only if connected to internet
			resp.Body.Close()
			fmt.Fprintf(GinkgoWriter, "helm executed\n")
			spec := helm.New("cockroachdb:3.0.8", "https://charts.helm.sh/stable")

			m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()}))
			Expect(m.MimeType()).To(Equal(helm2.ChartMediaType))
			defer Close(m)
			blob := Must(m.Reader())
			defer Close(blob)

			chart := Must(loader.LoadArchive(blob))
			Expect(chart.Name()).To(Equal("cockroachdb"))
			Expect(chart.Metadata.Version).To(Equal("3.0.8"))
		} else {
			fmt.Fprintf(GinkgoWriter, "helm test skipped\n")
		}
	})
})
