// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/artifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

////////////////////////////////////////////////////////////////////////////////
// manifestlaver

const (
	OCINAMESPACE = "ocm/value"
	OCIVERSION   = "v2.0"
	OCILAYER     = "manifestlayer"
)

var OCIDigests = common.Properties{
	"D_OCIMANIFEST1":     D_OCIMANIFEST1,
	"H_OCIARCHMANIFEST1": H_OCIARCHMANIFEST1,
	"D_OCIMANIFEST2":     D_OCIMANIFEST2,
	"H_OCIARCHMANIFEST2": H_OCIARCHMANIFEST2,
}

func OCIManifest1(env *builder.Builder) *artdesc.Descriptor {
	_, ldesc := OCIManifest1For(env, OCINAMESPACE, OCIVERSION)
	return ldesc
}

func OCIManifest1For(env *builder.Builder, ns, tag string) (*artdesc.Descriptor, *artdesc.Descriptor) {
	var ldesc *artdesc.Descriptor
	var mdesc *artdesc.Descriptor

	env.Namespace(ns, func() {
		mdesc = env.Manifest(tag, func() {
			env.Config(func() {
				env.BlobStringData(mime.MIME_JSON, "{}")
			})
			ldesc = env.Layer(func() {
				env.BlobStringData(mime.MIME_TEXT, OCILAYER)
			})
		})
	})
	return mdesc, ldesc
}

func OCIArtifactResource1(env *builder.Builder, name string, host string, funcs ...func()) {
	env.Resource(name, "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
		env.Access(
			ociartifact.New(oci.StandardOCIRef(host+".alias", OCINAMESPACE, OCIVERSION)),
		)
		env.Configure(funcs...)
	})
}

const (
	D_OCIMANIFEST1     = "0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f"
	H_OCIARCHMANIFEST1 = "b0692bcec00e0a875b6b280f3209d6776f3eca128adcb7e81e82fd32127c0c62"
)

var DS_OCIMANIFEST1 = &metav1.DigestSpec{
	HashAlgorithm:          sha256.Algorithm,
	NormalisationAlgorithm: artifact.OciArtifactDigestV1,
	Value:                  D_OCIMANIFEST1,
}

func HashManifest1(fmt string) string {
	// hash := "sha256:018520b2b249464a83e370619f544957b7936dd974468a128545eab88a0f53ed"
	hash := "xxx"
	if fmt == artifactset.FORMAT_OCI || fmt == artifactset.OCIArtifactSetDescriptorFileName {
		// hash = "sha256:334b587868e607fe2ce74c27d7f75e90b6391fe91b808b2d42ad1bfcc5651a66"
		// hash = "sha256:0a326cc646d24f48c9bc79d303f7626404d41f2646934ef713cd1917bd5480ce" // with gardener.cloud legacy format
		// hash = "sha256:fafabfc2f9861c2ecf0ee3fc584ef4fb92c927902c8f561f72542281097cff83"
		hash = "sha256:" + H_OCIARCHMANIFEST1
	}
	return hash
}

////////////////////////////////////////////////////////////////////////////////
// otherlayer

const (
	OCINAMESPACE2 = "ocm/ref"
	OCILAYER2     = "otherlayer"
)

var DS_OCIMANIFEST2 = &metav1.DigestSpec{
	HashAlgorithm:          sha256.Algorithm,
	NormalisationAlgorithm: artifact.OciArtifactDigestV1,
	Value:                  D_OCIMANIFEST2,
}

func OCIManifest2(env *builder.Builder) *artdesc.Descriptor {
	_, ldesc := OCIManifest2For(env, OCINAMESPACE2, OCIVERSION)
	return ldesc
}

func OCIManifest2For(env *builder.Builder, ns, tag string) (*artdesc.Descriptor, *artdesc.Descriptor) {
	var ldesc *artdesc.Descriptor
	var mdesc *artdesc.Descriptor

	env.Namespace(ns, func() {
		mdesc = env.Manifest(tag, func() {
			env.Config(func() {
				env.BlobStringData(mime.MIME_JSON, "{}")
			})
			ldesc = env.Layer(func() {
				env.BlobStringData(mime.MIME_TEXT, OCILAYER2)
			})
		})
	})
	return mdesc, ldesc
}

const (
	D_OCIMANIFEST2     = "c2d2dca275c33c1270dea6168a002d67c0e98780d7a54960758139ae19984bd7"
	H_OCIARCHMANIFEST2 = "cb85cd58b10e36343971691abbfe40200cb645c6e95f0bdabd111a30cf794708"
)

func HashManifest2(fmt string) string {
	// hash := "sha256:f6a519fb1d0c8cef5e8d7811911fc7cb170462bbce19d6df067dae041250de7f"
	hash := "xxx"
	if fmt == artifactset.FORMAT_OCI || fmt == artifactset.OCIArtifactSetDescriptorFileName {
		// hash = "sha256:253c2a52cd0e229ae97613b953e1aa5c0b8146ff653988904e858a676507d4f4"
		// hash = "sha256:d748056b98897e4894217daf2fed90c98d5603ca549256f0d9534994baee3795" // with gardener.cloud legacy format
		// hash = "sha256:e6b922b290aee4c9bca83d977b83dc3f91fe928e2085f0d45c1bde4544d3b19b"
		hash = "sha256:" + H_OCIARCHMANIFEST2
	}
	return hash
}

////////////////////////////////////////////////////////////////////////////////

const (
	OCIINDEXVERSION = "v2.0-index"
	OCINAMESPACE3   = "ocm/index"
)

func OCIIndex1(env *builder.Builder) *artdesc.Descriptor {
	var idesc *artdesc.Descriptor

	a1, _ := OCIManifest1For(env, OCINAMESPACE3, "")
	a2, _ := OCIManifest2For(env, OCINAMESPACE3, "")

	env.Namespace(OCINAMESPACE3, func() {
		idesc = env.Index(OCIINDEXVERSION, func() {
			env.Artifact(a1)
			env.Artifact(a2)
		})
	})
	return idesc
}
