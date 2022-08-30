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

package helm_test

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm/loader"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
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

func get(blob accessio.BlobAccess, expected []byte) []byte {
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

		artblob, err := helm.SynthesizeArtefactBlob("/testdata/testchart", env)
		Expect(err).To(Succeed())
		defer Close(artblob)
		set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, artblob)
		Expect(err).To(Succeed())
		defer Close(set)
		art, err := set.GetArtefact(set.GetMain().String())
		Expect(err).To(Succeed())
		defer Close(art)
		m := art.ManifestAccess().GetDescriptor()
		Expect(len(m.Layers)).To(Equal(2))

		config, err := art.ManifestAccess().GetConfigBlob()
		Expect(err).To(Succeed())
		get(config, meta)

		blob, err := set.GetBlob(m.Layers[1].Digest)
		Expect(err).To(Succeed())
		get(blob, prov)

		blob, err = set.GetBlob(m.Layers[0].Digest)
		Expect(err).To(Succeed())
		r, err := blob.Reader()
		Expect(err).To(Succeed())

		// unfortunately charts are not directly comparable, because of the order in the arrays AND the modified Chart.yaml
		found, err := loader.LoadArchive(r)
		Expect(err).To(Succeed())
		Expect(norm(found)).To(Equal(norm(chart)))
	})
})
