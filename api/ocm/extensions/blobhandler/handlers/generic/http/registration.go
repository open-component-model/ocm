package maven

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/registrations"
)

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler(BlobHandlerName, &RegistrationHandler{})
}

type Config struct {
	RepositoryURL string            `json:"url"`
	Path          string            `json:"path"`
	Method        string            `json:"method"`
	Credentials   CredentialsConfig `json:"credentials"`
}

type CredentialsConfig struct {
	Method string `json:"method"`
}

type rawConfig Config

func (c *Config) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &c.RepositoryURL)
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

type RegistrationHandler struct{}

var _ cpi.BlobHandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config cpi.BlobHandlerConfig, olist ...cpi.BlobHandlerOption) (bool, error) {
	if handler != "" {
		return true, fmt.Errorf("invalid %s handler %q", resourcetypes.HELM_CHART, handler)
	}
	if config == nil {
		return true, fmt.Errorf("http repository specification required")
	}
	cfg, err := registrations.DecodeConfig[Config](config)
	if err != nil {
		return true, errors.Wrapf(err, "blob handler configuration")
	}

	ctx.BlobHandlers().Register(NewArtifactHandler(cfg),
		cpi.ForArtifactType(resourcetypes.HELM_CHART),
		cpi.ForArtifactType(resourcetypes.BLOB),
		cpi.ForMimeType(mime.MIME_TGZ),
		cpi.ForMimeType(mime.MIME_TGZ_ALT),
		cpi.NewBlobHandlerOptions(olist...),
	)

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(_ cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("uploading http charts to http repositories", `
The <code>`+BlobHandlerName+`</code> uploader is able to upload artifacts to an arbitrary http path.
If registered the default mime type is: `+mime.MIME_TGZ+` and `+mime.MIME_TGZ_ALT+`.

It accepts a plain string for the URL and a config with the following field:
'url': the URL of the http server to request.
'method': the HTTP Method to use for the request for an individual artifact.
''
`)
}
