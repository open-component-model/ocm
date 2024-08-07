package ocirepo

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/api/utils/registrations"
)

const PATH = "oci/artifact"

func init() {
	download.RegisterHandlerRegistrationHandler(PATH, &RegistrationHandler{})
}

var supportedMimeTypes = []string{
	artifactset.MediaType(artdesc.MediaTypeImageManifest),
	artifactset.MediaType(artdesc.MediaTypeImageIndex),
}

type Config = ociuploadattr.Attribute

func AttributeDescription() map[string]string {
	return ociuploadattr.AttributeDescription()
}

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
	if opts.MimeType != "" && !slices.Contains(supportedMimeTypes, opts.MimeType) {
		return true, errors.Wrapf(err, "mime type %s not supported", opts.MimeType)
	}

	h := New(attr)
	if opts.MimeType == "" {
		for _, m := range supportedMimeTypes {
			opts.MimeType = m
			download.For(ctx).Register(h, opts)
		}
	} else {
		download.For(ctx).Register(h, opts)
	}

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(ctx cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("uploading an OCI artifact to an OCI registry", `
The <code>artifact</code> downloader is able to transfer OCI artifact-like resources
into an OCI registry given by the combination of the download target and the
registration config.

If no config is given, the target must be an OCI reference with a potentially
omitted repository. The repo part is derived from the reference hint provided
by the resource's access specification.

If the config is given, the target is used as repository name prefixed with an
optional repository prefix given by the configuration.

The following artifact media types are supported:
`+listformat.FormatList("", supportedMimeTypes...)+`
It accepts a config with the following fields:
`+listformat.FormatMapElements("", AttributeDescription()),
	)
}
