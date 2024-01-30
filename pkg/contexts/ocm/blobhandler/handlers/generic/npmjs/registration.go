package npmjs

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/registrations"
)

type Config struct {
	Url string `json:"url"`
}

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler("ocm/npmPackage", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ cpi.BlobHandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config cpi.BlobHandlerConfig, olist ...cpi.BlobHandlerOption) (bool, error) {
	if handler != "" {
		return true, fmt.Errorf("invalid npmjsArtifact handler %q", handler)
	}
	if config == nil {
		return true, fmt.Errorf("npmjs target specification required")
	}
	cfg, err := registrations.DecodeConfig[Config](config)
	if err != nil {
		return true, errors.Wrapf(err, "blob handler configuration")
	}

	ctx.BlobHandlers().Register(NewArtifactHandler(cfg),
		cpi.ForArtifactType(resourcetypes.NPM_PACKAGE),
		cpi.ForMimeType(mime.MIME_TGZ),
		cpi.NewBlobHandlerOptions(olist...),
	)

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(ctx cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("uploading npmjs artifacts", `
The <code>npmjsArtifacts</code> uploader is able to upload npmjs artifacts
as artifact archive according to the npmjs package spec.
The following artifact media types are supported: `+mime.MIME_TGZ+`
By default, it is registered for these mimetypes.

It accepts a config with the following fields:
'url': the URL of the npmjs repository.
If not given, the default npmjs.com repository.
`,
	)
}
