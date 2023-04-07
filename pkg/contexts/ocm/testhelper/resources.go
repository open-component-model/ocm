// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"github.com/open-component-model/ocm/pkg/common"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/blob"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

func TextResourceDigestSpec(d string) *metav1.DigestSpec {
	return &metav1.DigestSpec{
		HashAlgorithm:          sha256.Algorithm,
		NormalisationAlgorithm: blob.GenericBlobDigestV1,
		Value:                  d,
	}
}

var Digests = common.Properties{
	"D_TESTDATA":  D_TESTDATA,
	"D_OTHERDATA": D_OTHERDATA,
}

const D_TESTDATA = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"

var DS_TESTDATA = TextResourceDigestSpec(D_TESTDATA)

func TestDataResource(env *builder.Builder) {
	env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
		env.BlobStringData(mime.MIME_TEXT, "testdata")
	})
}

const D_OTHERDATA = "54b8007913ec5a907ca69001d59518acfd106f7b02f892eabf9cae3f8b2414b4"

var DS_OTHERDATA = TextResourceDigestSpec(D_OTHERDATA)

func OtherDataResource(env *builder.Builder) {
	env.Resource("otherdata", "", "PlainText", metav1.LocalRelation, func() {
		env.BlobStringData(mime.MIME_TEXT, "otherdata")
	})
}
