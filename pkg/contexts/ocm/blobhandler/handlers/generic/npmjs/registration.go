package npmjs

import (
	"encoding/json"
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

type rawConfig Config

func (c *Config) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &c.Url)
	if err == nil {
		return nil
	}
	var raw rawConfig
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	*c = Config(raw)

	return nil
}

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler(BLOB_HANDLER_NAME, &RegistrationHandler{})
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
The <code>`+BLOB_HANDLER_NAME+`</code> uploader is able to upload npmjs artifacts
as artifact archive according to the npmjs package spec.
If registered the default mime type is: `+mime.MIME_TGZ+`

It accepts a plain string for the URL or a config with the following field:
'url': the URL of the npmjs repository.
`,
	)
}
