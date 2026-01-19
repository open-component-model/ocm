package helm

import (
	"fmt"
	"slices"

	"github.com/mandelsoft/goutils/errors"
	helmregistry "helm.sh/helm/v4/pkg/registry"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/cpi"
	downloadhandlers "ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/api/utils/registrations"
)

const PATH = "helm/artifact"

func init() {
	downloadhandlers.RegisterHandlerRegistrationHandler(PATH, &RegistrationHandler{})
}

var supportedMimeTypes = []string{
	artifactset.MediaType(artdesc.MediaTypeImageManifest),
	helmregistry.ChartLayerMediaType,
}

type RegistrationHandler struct{}

var _ downloadhandlers.HandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx downloadhandlers.Target, config downloadhandlers.HandlerConfig, olist ...downloadhandlers.HandlerOption) (bool, error) {
	var err error

	if handler != "" {
		return true, fmt.Errorf("invalid helm handler %q", handler)
	}

	if config != nil {
		return true, fmt.Errorf("helm downloader does not support configuration")
	}

	opts := downloadhandlers.NewHandlerOptions(olist...)
	if opts.MimeType != "" && !slices.Contains(supportedMimeTypes, opts.MimeType) {
		return true, errors.Wrapf(err, "mime type %s not supported", opts.MimeType)
	}

	h := New()
	if opts.MimeType == "" {
		for _, m := range supportedMimeTypes {
			opts.MimeType = m
			downloadhandlers.For(ctx).Register(h, opts)
		}
	} else {
		downloadhandlers.For(ctx).Register(h, opts)
	}

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(ctx cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo(`download helm chart
resources`, `
The <code>helm</code> downloader is able to download helm chart resources as 
helm chart packages. Thus, the downloader may perform transformations. 
For example, if the helm chart is currently stored as an oci artifact, the 
downloader performs the necessary extraction to provide the helm chart package 
from within that oci artifact.

The following artifact media types are supported:
`+listformat.FormatList("", supportedMimeTypes...)+`
It accepts no config.`)
}
