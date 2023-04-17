// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc_test

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/normalizations"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	compdesc "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

var CD1 = `
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
      labels:
      - name: label1
        value: foo
      - name: label2
        value: bar
        signing: true
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
`

var CD2 = `
  component:
    componentReferences: []
    name: github.com/vasu1124/introspect
    provider: internal
    repositoryContexts:
    - baseUrl: ghcr.io/vasu1124/ocm
      componentNameMapping: urlPath
      type: ociRegistry
    - baseUrl: ghcr.io
      componentNameMapping: urlPath
      subPath: mandelsoft/cnudie
      type: OCIRegistry
    resources:
    - access:
        globalAccess:
          digest: sha256:7f0168496f273c1e2095703a050128114d339c580b0906cd124a93b66ae471e2
          mediaType: application/vnd.docker.distribution.manifest.v2+tar+gzip
          ref: ghcr.io/mandelsoft/cnudie/component-descriptors/github.com/vasu1124/introspect
          size: 29047129
          type: ociBlob
        localReference: sha256:7f0168496f273c1e2095703a050128114d339c580b0906cd124a93b66ae471e2
        mediaType: application/vnd.docker.distribution.manifest.v2+tar+gzip
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: ociArtifactDigest/v1
        value: 6a1c7637a528ab5957ab60edf73b5298a0a03de02a96be0313ee89b22544840c
      labels:
      - name: label1
        value: foo
      - name: label2
        value: bar
        signing: true
      name: introspect-image
      relation: local
      type: ociImage
      version: 1.0.0
    - access:
        globalAccess:
          digest: sha256:d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf
          ref: ghcr.io/mandelsoft/cnudie/component-descriptors/github.com/vasu1124/introspect
          size: 632
          type: ociBlob
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
        imageReference: ghcr.io/mandelsoft/cnudie/vasu1124/helm/introspect-helm:0.1.0
        type: ociRegistry
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
`

var _ = Describe("Normalization", func() {
	var cd1 *compdesc.ComponentDescriptor
	var cd2 *compdesc.ComponentDescriptor

	BeforeEach(func() {
		var err error
		cd1, err = compdesc.Decode([]byte(CD1))
		Expect(err).To(Succeed())
		cd2, err = compdesc.Decode([]byte(CD2))
		Expect(err).To(Succeed())
	})

	It("hashes first", func() {
		n, err := compdesc.Normalize(cd1, compdesc.JsonNormalisationV1)
		Expect(err).To(Succeed())
		Expect(string(n)).To(Equal("[{\"component\":[{\"componentReferences\":[]},{\"name\":\"github.com/vasu1124/introspect\"},{\"provider\":\"internal\"},{\"resources\":[[{\"digest\":[{\"hashAlgorithm\":\"SHA-256\"},{\"normalisationAlgorithm\":\"ociArtifactDigest/v1\"},{\"value\":\"6a1c7637a528ab5957ab60edf73b5298a0a03de02a96be0313ee89b22544840c\"}]},{\"labels\":[[{\"name\":\"label2\"},{\"signing\":true},{\"value\":\"bar\"}]]},{\"name\":\"introspect-image\"},{\"relation\":\"local\"},{\"type\":\"ociImage\"},{\"version\":\"1.0.0\"}],[{\"digest\":[{\"hashAlgorithm\":\"SHA-256\"},{\"normalisationAlgorithm\":\"genericBlobDigest/v1\"},{\"value\":\"d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf\"}]},{\"name\":\"introspect-blueprint\"},{\"relation\":\"local\"},{\"type\":\"landscaper.gardener.cloud/blueprint\"},{\"version\":\"1.0.0\"}],[{\"digest\":[{\"hashAlgorithm\":\"SHA-256\"},{\"normalisationAlgorithm\":\"ociArtifactDigest/v1\"},{\"value\":\"6229be2be7e328f74ba595d93b814b590b1aa262a1b85e49cc1492795a9e564c\"}]},{\"name\":\"introspect-helm\"},{\"relation\":\"external\"},{\"type\":\"helm\"},{\"version\":\"0.1.0\"}]]},{\"version\":\"1.0.0\"}]},{\"meta\":[{\"schemaVersion\":\"v2\"}]}]"))
		o, err := compdesc.Normalize(cd2, compdesc.JsonNormalisationV1)
		Expect(err).To(Succeed())
		Expect(o).To(Equal(n))
	})

	It("hashes v2", func() {
		n, err := compdesc.Normalize(cd1, compdesc.JsonNormalisationV2)
		Expect(err).To(Succeed())
		// Expect(string(n)).To(Equal("[{\"component\":[{\"componentReferences\":[]},{\"name\":\"github.com/vasu1124/introspect\"},{\"provider\":[{\"name\":\"internal\"}]},{\"resources\":[[{\"digest\":[{\"hashAlgorithm\":\"SHA-256\"},{\"normalisationAlgorithm\":\"ociArtifactDigest/v1\"},{\"value\":\"6a1c7637a528ab5957ab60edf73b5298a0a03de02a96be0313ee89b22544840c\"}]},{\"labels\":[[{\"name\":\"label2\"},{\"signing\":true},{\"value\":\"bar\"}]]},{\"name\":\"introspect-image\"},{\"relation\":\"local\"},{\"type\":\"ociImage\"},{\"version\":\"1.0.0\"}],[{\"digest\":[{\"hashAlgorithm\":\"SHA-256\"},{\"normalisationAlgorithm\":\"genericBlobDigest/v1\"},{\"value\":\"d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf\"}]},{\"name\":\"introspect-blueprint\"},{\"relation\":\"local\"},{\"type\":\"landscaper.gardener.cloud/blueprint\"},{\"version\":\"1.0.0\"}],[{\"digest\":[{\"hashAlgorithm\":\"SHA-256\"},{\"normalisationAlgorithm\":\"ociArtifactDigest/v1\"},{\"value\":\"6229be2be7e328f74ba595d93b814b590b1aa262a1b85e49cc1492795a9e564c\"}]},{\"name\":\"introspect-helm\"},{\"relation\":\"external\"},{\"type\":\"helm\"},{\"version\":\"0.1.0\"}]]},{\"sources\":[[{\"name\":\"introspect\"},{\"type\":\"git\"},{\"version\":\"1.0.0\"}]]},{\"version\":\"1.0.0\"}]}]"))
		Expect(string(n)).To(Equal(`{"component":{"componentReferences":[],"name":"github.com/vasu1124/introspect","provider":{"name":"internal"},"resources":[{"digest":{"hashAlgorithm":"SHA-256","normalisationAlgorithm":"ociArtifactDigest/v1","value":"6a1c7637a528ab5957ab60edf73b5298a0a03de02a96be0313ee89b22544840c"},"labels":[{"name":"label2","signing":true,"value":"bar"}],"name":"introspect-image","relation":"local","type":"ociImage","version":"1.0.0"},{"digest":{"hashAlgorithm":"SHA-256","normalisationAlgorithm":"genericBlobDigest/v1","value":"d1187ac17793b2f5fa26175c21cabb6ce388871ae989e16ff9a38bd6b32507bf"},"name":"introspect-blueprint","relation":"local","type":"landscaper.gardener.cloud/blueprint","version":"1.0.0"},{"digest":{"hashAlgorithm":"SHA-256","normalisationAlgorithm":"ociArtifactDigest/v1","value":"6229be2be7e328f74ba595d93b814b590b1aa262a1b85e49cc1492795a9e564c"},"name":"introspect-helm","relation":"external","type":"helm","version":"0.1.0"}],"sources":[{"name":"introspect","type":"git","version":"1.0.0"}],"version":"1.0.0"}}`))
	})
})
