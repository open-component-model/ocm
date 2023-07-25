// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm_test

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/common/accessobj"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/v2/pkg/env"
	"github.com/open-component-model/ocm/v2/pkg/env/builder"
	"github.com/open-component-model/ocm/v2/pkg/helm/loader"
)

type Files []*chart.File

var _ sort.Interface = (Files)(nil)

func (f Files) Len() int {
	return len(f)
}

func (f Files) Less(i, j int) bool {
	return strings.Compare(f[i].Name, f[j].Name) < 0
}

func (f Files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func norm(chart *chart.Chart) *chart.Chart {
	dir, err := os.MkdirTemp("", "helmchart-")
	Expect(err).To(Succeed())
	defer os.RemoveAll(dir)

	path, err := chartutil.Save(chart, dir)
	Expect(err).To(Succeed())
	chart, err = loader.Load(path, osfs.New())
	Expect(err).To(Succeed())
	//	sort.Sort(Files(chart.Raw))
	//	sort.Sort(Files(chart.Files))
	//	sort.Sort(Files(chart.Templates))
	return chart
}

func get(blob accessio.DataAccess, expected []byte) []byte {
	data, err := blob.Get()
	ExpectWithOffset(1, err).To(Succeed())
	if expected != nil {
		ExpectWithOffset(1, string(data)).To(Equal(string(expected)))
	}
	return data
}

var _ = Describe("art parsing", func() {
	It("succeeds", func() {
		env := builder.NewBuilder(env.NewEnvironment(env.TestData()))
		defer vfs.Cleanup(env)

		prov, err := env.ReadFile("/testdata/testchart.prov")
		Expect(err).To(Succeed())
		chart, err := loader.Load("/testdata/testchart", env)
		Expect(err).To(Succeed())
		meta, err := json.Marshal(chart.Metadata)
		Expect(err).To(Succeed())

		artblob, err := helm.SynthesizeArtifactBlob(loader.VFSLoader("/testdata/testchart", env))
		Expect(err).To(Succeed())
		defer Close(artblob)
		set, err := artifactset.OpenFromBlob(accessobj.ACC_READONLY, artblob)
		Expect(err).To(Succeed())
		defer Close(set)
		art, err := set.GetArtifact(set.GetMain().String())
		Expect(err).To(Succeed())
		defer Close(art)

		ma := art.ManifestAccess()
		m := ma.GetDescriptor()
		Expect(len(m.Layers)).To(Equal(2))

		config, err := art.ManifestAccess().GetConfigBlob()
		Expect(err).To(Succeed())
		get(config, meta)

		_, data, err := set.GetBlobData(m.Layers[1].Digest)
		Expect(err).To(Succeed())
		get(data, prov)

		_, data, err = set.GetBlobData(m.Layers[0].Digest)
		Expect(err).To(Succeed())
		r, err := data.Reader()
		Expect(err).To(Succeed())

		blob, err := ma.GetBlob(m.Layers[1].Digest)
		Expect(err).To(Succeed())
		get(blob, prov)

		// unfortunately charts are not directly comparable, because of the order in the arrays AND the modified Chart.yaml
		found, err := loader.LoadArchive(r)
		Expect(err).To(Succeed())
		Expect(norm(found)).To(Equal(norm(chart)))
	})
})
