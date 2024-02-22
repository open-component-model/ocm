// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/go-test/deep"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/defaultmerge"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("Component Descripor Copy Test Suitet", func() {
	Context("compdesc copy", func() {
		It("copies CD", func() {

			labels := v1.Labels{
				*Must(v1.NewLabel("label", "value",
					v1.WithVersion("v1"),
					v1.WithSigning(true),
					v1.WithMerging(defaultmerge.ALGORITHM, defaultmerge.NewConfig(defaultmerge.MODE_LOCAL)))),
			}
			cd := compdesc.New("mandelsoft.org/test", "1.0.0")
			cd.Metadata.ConfiguredVersion = "xxx"
			cd.ObjectMeta.CreationTime = compdesc.NewTimestampP()
			cd.ObjectMeta.Provider = v1.Provider{
				Name:   "mandelsoft",
				Labels: labels,
			}
			cd.ObjectMeta.Labels = labels
			cd.RepositoryContexts = runtime.UnstructuredTypedObjectList{
				runtime.NewEmptyUnstructured("repo"),
			}
			cd.Resources = compdesc.Resources{
				compdesc.Resource{
					ResourceMeta: compdesc.ResourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:          "resc1",
							Version:       "v1",
							ExtraIdentity: v1.NewExtraIdentity("id", "a"),
							Labels:        labels,
						},
						Type:       "rsc",
						Relation:   v1.LocalRelation,
						SourceRefs: nil,
						Digest: &v1.DigestSpec{
							HashAlgorithm:          "hashalgo",
							NormalisationAlgorithm: "normalgo",
							Value:                  "digest",
						},
					},
					Access: ociartifact.New("oci.com/image"),
				},
			}
			cd.Sources = compdesc.Sources{
				compdesc.Source{
					SourceMeta: compdesc.SourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:          "src1",
							Version:       "v2",
							ExtraIdentity: v1.NewExtraIdentity("id", "b"),
							Labels:        labels,
						},
						Type: "src",
					},
					Access: ociartifact.New("oci.com/otherimage"),
				},
			}
			cd.References = compdesc.References{
				compdesc.ComponentReference{
					ElementMeta:   compdesc.ElementMeta{},
					ComponentName: "",
					Digest:        nil,
				},
			}

			cd.Signatures = v1.Signatures{
				v1.Signature{
					Name: "sig",
					Digest: v1.DigestSpec{
						HashAlgorithm:          "hashalgo2",
						NormalisationAlgorithm: "normalgo2",
						Value:                  "digest2",
					},
					Signature: v1.SignatureSpec{
						Algorithm: "sigalgo",
						Value:     "sig",
						MediaType: "media",
						Issuer:    "issuer",
					},
					Timestamp: &v1.TimestampSpec{
						Value: "ts",
						Time:  compdesc.NewTimestampP(),
					},
				},
			}
			cp := cd.Copy()

			Expect(deep.Equal(cd, cp)).To(BeNil())
		})
	})
})
