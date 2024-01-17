// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package directory

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/wget"
	"github.com/open-component-model/ocm/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/blobaccess"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`
	// URL defines the url from which the artifact is downloaded.
	URL string `json:"url"`
	// MimeType defines the mime type of the artifact.
	MimeType string `json:"mediaType"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(url, mimeType string) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		URL:      url,
		MimeType: mimeType,
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	var allErrs field.ErrorList
	if s.URL == "" {
		pathField := fldPath.Child("URL")
		allErrs = append(allErrs, field.Invalid(pathField, s.URL, "no url"))
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	access, err := wget.BlobAccessForWget(s.URL,
		wget.WithCredentialContext(ctx),
		wget.WithLoggingContext(ctx),
		wget.WithMimeType(s.MimeType),
	)
	return access, "", err
}
