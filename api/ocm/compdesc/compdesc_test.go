package compdesc_test

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	compdescv3 "ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1"
)

func NormalizeYAML(y string) string {
	var t map[string]interface{}
	err := compdesc.DefaultYAMLCodec.Decode([]byte(y), &t)
	Expect(err).To(Succeed())
	d, err := compdesc.DefaultYAMLCodec.Encode(t)
	Expect(err).To(Succeed())
	return string(d)
}

var _ = Describe("serialization", func() {
	CDv2 := NormalizeYAML(`
  component:
    componentReferences: []
    name: github.com/vasu1124/introspect
    provider: internal
    repositoryContexts:
    - baseUrl: ghcr.io/vasu1124/ocm
      componentNameMapping: urlPath
      type: ociRegistry
    resources:
    - access:
        localReference: sha256:7f0168496f273c1e2095703a050128114d339c580b0906cd124a93b66ae471e2
        mediaType: application/vnd.docker.distribution.manifest.v2+tar+gzip
        referenceName: vasu1124/introspect:1.0.0
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: ociArtifactDigest/v1
        value: 6a1c7637a528ab5957ab60edf73b5298a0a03de02a96be0313ee89b22544840c
      name: introspect-image
      relation: local
      type: ociImage
      version: 1.0.0
    - access:
        localReference: sha256:d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf
        mediaType: ""
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: genericBlobDigest/v1
        value: d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf
      name: introspect-blueprint
      relation: local
      type: landscaper.gardener.cloud/blueprint
      version: 1.0.0
    - access:
        localReference: sha256:4186663939459149a21c0bb1cd7b8ff86e0021b29ca45069446d046f808e6bfe
        mediaType: application/vnd.oci.image.manifest.v1+tar+gzip
        referenceName: vasu1124/helm/introspect-helm:0.1.0
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: ociArtifactDigest/v1
        value: 6229be2be7e328f74ba595d93b814b590b1aa262a1b85e49cc1492795a9e564c
      name: introspect-helm
      relation: external
      type: helm
      version: 0.1.0
    sources:
    - access:
        repository: github.com/vasu1124/introspect
        type: git
      name: introspect
      type: git
      version: 1.0.0
    version: 1.0.0
  meta:
    schemaVersion: v2 
`)

	CDv3 := NormalizeYAML(fmt.Sprintf(`
apiVersion: %s
kind: ComponentVersion
metadata:
  name: github.com/vasu1124/introspect
  provider:
    name: internal
  version: 1.0.0
repositoryContexts:
- baseUrl: ghcr.io/vasu1124/ocm
  componentNameMapping: urlPath
  type: ociRegistry
spec:
  resources:
  - access:
      localReference: sha256:7f0168496f273c1e2095703a050128114d339c580b0906cd124a93b66ae471e2
      mediaType: application/vnd.docker.distribution.manifest.v2+tar+gzip
      referenceName: vasu1124/introspect:1.0.0
      type: localBlob
    digest:
      hashAlgorithm: SHA-256
      normalisationAlgorithm: ociArtifactDigest/v1
      value: 6a1c7637a528ab5957ab60edf73b5298a0a03de02a96be0313ee89b22544840c
    name: introspect-image
    relation: local
    type: ociImage
    version: 1.0.0
  - access:
      localReference: sha256:d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf
      mediaType: ""
      type: localBlob
    digest:
      hashAlgorithm: SHA-256
      normalisationAlgorithm: genericBlobDigest/v1
      value: d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf
    name: introspect-blueprint
    relation: local
    type: landscaper.gardener.cloud/blueprint
    version: 1.0.0
  - access:
      localReference: sha256:4186663939459149a21c0bb1cd7b8ff86e0021b29ca45069446d046f808e6bfe
      mediaType: application/vnd.oci.image.manifest.v1+tar+gzip
      referenceName: vasu1124/helm/introspect-helm:0.1.0
      type: localBlob
    digest:
      hashAlgorithm: SHA-256
      normalisationAlgorithm: ociArtifactDigest/v1
      value: 6229be2be7e328f74ba595d93b814b590b1aa262a1b85e49cc1492795a9e564c
    name: introspect-helm
    relation: external
    type: helm
    version: 0.1.0
  sources:
  - access:
      repository: github.com/vasu1124/introspect
      type: git
    name: introspect
    type: git
    version: 1.0.0
`, compdescv3.SchemaVersion))

	It("deserializes v2", func() {
		cd, err := compdesc.Decode([]byte(CDv2))
		Expect(err).To(Succeed())

		data, err := compdesc.Encode(cd)
		Expect(err).To(Succeed())

		Expect(string(data)).To(Equal(CDv2))
	})

	It("deserializes "+compdescv3.VersionName, func() {
		cd, err := compdesc.Decode([]byte(CDv2))
		Expect(err).To(Succeed())

		cd.Metadata.ConfiguredVersion = compdescv3.SchemaVersion
		data, err := compdesc.Encode(cd)
		Expect(err).To(Succeed())
		Expect(string(data)).To(StringEqualWithContext(CDv3))
		cd2, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd2).To(Equal(cd))
	})
})
