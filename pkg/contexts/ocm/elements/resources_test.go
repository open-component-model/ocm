package elements_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/blob"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/elements"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

type value struct {
	Field string `json:"field"`
}

var _ = Describe("resources", func() {
	It("configures resource meta", func() {
		m := Must(me.ResourceMeta("name", "type",
			me.WithVersion("v1"),
			me.WithExtraIdentity("extra", "value"),
			me.WithLabel("label", value{"value"}, metav1.WithSigning(), metav1.WithVersion("v1")),
			me.WithSourceRef("name", "image").WithLabel("prop", "x"),
			me.WithDigest(sha256.Algorithm, blob.GenericBlobDigestV1, "0815"),
		))
		Expect(m).To(YAMLEqual(`
name: name
type: type
version: v1
relation: local
extraIdentity:
  extra: value
labels:
  - name: label
    version: v1
    value:
      field: value
    signing: true
srcRefs:
  - identitySelector:
      name: image
    labels:
    - name: prop
      value: x
digest:
    hashAlgorithm: SHA-256
    normalisationAlgorithm: genericBlobDigest/v1
    value: "0815"
`))
	})
})
