package wget

import (
	"bytes"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/wget"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`
	// URL defines the url from which the artifact is downloaded.
	URL string `json:"url"`
	// MimeType defines the mime type of the artifact.
	MimeType string `json:"mediaType"`
	// Header to be passed in the http request
	Header map[string][]string `json:"header"`
	// Verb is the http verb to be used for the request
	Verb string `json:"verb"`
	// Body is the body to be included in the http request
	Body string `json:"body"`
	// NoRedirect allows to disable redirects
	NoRedirect bool `json:"noRedirect"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(url, mimeType string, header map[string][]string, verb string, body string, noRedirect bool) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		URL:        url,
		MimeType:   mimeType,
		Header:     header,
		Verb:       verb,
		Body:       body,
		NoRedirect: noRedirect,
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
	access, err := wget.BlobAccess(s.URL,
		wget.WithCredentialContext(ctx),
		wget.WithLoggingContext(ctx),
		wget.WithMimeType(s.MimeType),
		wget.WithHeader(s.Header),
		wget.WithVerb(s.Verb),
		wget.WithBody(bytes.NewReader([]byte(s.Body))),
		wget.WithNoRedirect(s.NoRedirect),
	)
	return access, "", err
}
