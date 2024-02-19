package ocirepo

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/registrations"
)

const PATH = "oci/artifact"

func init() {
	download.RegisterHandlerRegistrationHandler(PATH, &RegistrationHandler{})
}

type Config = ociuploadattr.Attribute

type RegistrationHandler struct{}

var _ download.HandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx download.Target, config download.HandlerConfig, olist ...download.HandlerOption) (bool, error) {
	var err error

	if handler != "" {
		return true, fmt.Errorf("invalid ocireg handler %q", handler)
	}

	attr, err := registrations.DecodeConfig[Config](config, ociuploadattr.AttributeType{}.Decode)
	if err != nil {
		return true, errors.Wrapf(err, "cannot unmarshal download handler configuration")
	}

	opts := download.NewHandlerOptions(olist...)
	if opts.MimeType != "" && !slices.Contains(artdesc.SupportedMimeTypes, opts.MimeType) {
		return true, errors.Wrapf(err, "mime type %s not supported", opts.MimeType)
	}

	h := New(attr)
	if opts.MimeType == "" {
		for _, m := range artdesc.SupportedMimeTypes {
			opts.MimeType = m
			download.For(ctx).Register(h, opts)
		}
	} else {
		download.For(ctx).Register(h, opts)
	}

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(_ cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("download OCI artifacts", `
The <code>`+PATH+`</code> downloader is able to download OCI artifacts
as artifact archive according to the OCI distribution spec.
The following artifact media types are supported:
`+listformat.FormatList("", artdesc.ArchiveBlobTypes()...)+`
By default, it is registered for these mimetypes.

It accepts a config with the following fields:
`+listformat.FormatMapElements("", ociuploadattr.AttributeDescription())+`
Alternatively, a single string value can be given representing an OCI repository
reference.`,
	)
}
