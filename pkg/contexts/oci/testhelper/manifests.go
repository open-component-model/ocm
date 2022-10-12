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

package testhelper

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
)

////////////////////////////////////////////////////////////////////////////////
// manifestlaver

const (
	OCINAMESPACE = "ocm/value"
	OCIVERSION   = "v2.0"
)

func OCIManifest1(env *builder.Builder) *artdesc.Descriptor {
	var ldesc *artdesc.Descriptor

	env.Namespace(OCINAMESPACE, func() {
		env.Manifest(OCIVERSION, func() {
			env.Config(func() {
				env.BlobStringData(mime.MIME_JSON, "{}")
			})
			ldesc = env.Layer(func() {
				env.BlobStringData(mime.MIME_TEXT, "manifestlayer")
			})
		})
	})
	return ldesc
}

func HashManifest1(fmt string) string {
	// hash := "sha256:018520b2b249464a83e370619f544957b7936dd974468a128545eab88a0f53ed"
	hash := "xxx"
	if fmt == artefactset.FORMAT_OCI || fmt == artefactset.OCIArtefactSetDescriptorFileName {
		// hash = "sha256:334b587868e607fe2ce74c27d7f75e90b6391fe91b808b2d42ad1bfcc5651a66"
		hash = "sha256:0a326cc646d24f48c9bc79d303f7626404d41f2646934ef713cd1917bd5480ce"
	}
	return hash
}

////////////////////////////////////////////////////////////////////////////////
// otherlayer

const OCINAMESPACE2 = "ocm/ref"

func OCIManifest2(env *builder.Builder) *artdesc.Descriptor {
	var ldesc *artdesc.Descriptor

	env.Namespace(OCINAMESPACE2, func() {
		env.Manifest(OCIVERSION, func() {
			env.Config(func() {
				env.BlobStringData(mime.MIME_JSON, "{}")
			})
			ldesc = env.Layer(func() {
				env.BlobStringData(mime.MIME_TEXT, "otherlayer")
			})
		})
	})
	return ldesc
}

func HashManifest2(fmt string) string {
	// hash := "sha256:f6a519fb1d0c8cef5e8d7811911fc7cb170462bbce19d6df067dae041250de7f"
	hash := "xxx"
	if fmt == artefactset.FORMAT_OCI || fmt == artefactset.OCIArtefactSetDescriptorFileName {
		// hash = "sha256:253c2a52cd0e229ae97613b953e1aa5c0b8146ff653988904e858a676507d4f4"
		hash = "sha256:d748056b98897e4894217daf2fed90c98d5603ca549256f0d9534994baee3795"
	}
	return hash
}
