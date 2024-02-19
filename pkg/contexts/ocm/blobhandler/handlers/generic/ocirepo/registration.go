package ocirepo

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/registrations"
)

type Config = ociuploadattr.Attribute

const UPLOADER_NAME = "ocm/ociArtifacts"

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler(UPLOADER_NAME, &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ cpi.BlobHandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config cpi.BlobHandlerConfig, olist ...cpi.BlobHandlerOption) (bool, error) {
	if handler != "" {
		return true, fmt.Errorf("invalid ociArtifact handler %q", handler)
	}
	if config == nil {
		return true, fmt.Errorf("oci target specification required")
	}
	attr, err := registrations.DecodeConfig[Config](config, ociuploadattr.AttributeType{}.Decode)
	if err != nil {
		return true, errors.Wrapf(err, "blob handler configuration")
	}

	var mimes []string
	opts := cpi.NewBlobHandlerOptions(olist...)
	if opts.MimeType != "" {
		found := false
		for _, a := range artdesc.ArchiveBlobTypes() {
			if a == opts.MimeType {
				found = true
				break
			}
		}
		if !found {
			return true, fmt.Errorf("unexpected type mime type %q for oci blob handler target", opts.MimeType)
		}
		mimes = append(mimes, opts.MimeType)
	} else {
		mimes = artdesc.ArchiveBlobTypes()
	}

	h := NewArtifactHandler(attr)
	for _, m := range mimes {
		opts.MimeType = m
		ctx.BlobHandlers().Register(h, opts)
	}

	return true, nil
}

func AttributeDescription() map[string]string {
	return ociuploadattr.AttributeDescription()
}

func (r *RegistrationHandler) GetHandlers(_ cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("upload an OCI artifact to an OCI registry", `
The <code>`+UPLOADER_NAME+`</code> uploader is able to transfer OCI artifact-like resources
into an OCI registry given by the combination of the upload target and the registration config.

If no config is given, the target must be an OCI reference with a potentially
omitted repository. The repo part is derived from the reference hint provided
by the resource's access specification.

If the config is given, the target is used as repository name prefixed with an
optional repository prefix given by the configuration.

The following artifact media types are supported:
`+listformat.FormatList("", artifactset.SupportedMimeTypes...)+`
It accepts a config with the following fields:
`+listformat.FormatMapElements("", AttributeDescription()),
	)
}
