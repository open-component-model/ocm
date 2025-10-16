package maven

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/registrations"
)

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler(BlobHandlerName, &RegistrationHandler{})
}

type Config struct {
	Url        string         `json:"url"`
	Path       string         `json:"path"`
	FileSystem vfs.FileSystem `json:"-"`
}

func NewFileConfig(path string, fss ...vfs.FileSystem) *Config {
	return &Config{
		Path:       path,
		FileSystem: utils.FileSystem(fss...),
	}
}

func NewUrlConfig(url string, fss ...vfs.FileSystem) *Config {
	return &Config{
		Url:        url,
		FileSystem: utils.FileSystem(fss...),
	}
}

type rawConfig Config

func (c *Config) GetRepository(ctx cpi.ContextProvider) (*maven.Repository, error) {
	if c.Url != "" && c.Path != "" {
		return nil, fmt.Errorf("cannot specify both url and path")
	}
	if c.Url != "" {
		return maven.NewUrlRepository(c.Url, general.OptionalDefaulted(vfsattr.Get(ctx.OCMContext()), c.FileSystem))
	}
	if c.Path != "" {
		return maven.NewFileRepository(c.Path, general.OptionalDefaulted(vfsattr.Get(ctx.OCMContext()), c.FileSystem)), nil
	}
	return nil, fmt.Errorf("must specify either url or path")
}

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

type RegistrationHandler struct{}

var _ cpi.BlobHandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config cpi.BlobHandlerConfig, olist ...cpi.BlobHandlerOption) (bool, error) {
	if handler != "" {
		return true, fmt.Errorf("invalid %s handler %q", resourcetypes.MAVEN_PACKAGE, handler)
	}
	if config == nil {
		return true, fmt.Errorf("maven target specification required")
	}
	cfg, err := registrations.DecodeConfig[Config](config)
	if err != nil {
		return true, errors.Wrapf(err, "blob handler configuration")
	}

	ctx.BlobHandlers().Register(NewArtifactHandler(cfg),
		cpi.ForArtifactType(resourcetypes.MAVEN_PACKAGE),
		cpi.ForMimeType(mime.MIME_TGZ),
		cpi.NewBlobHandlerOptions(olist...),
	)

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(_ cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("uploading maven artifacts", `
The <code>`+BlobHandlerName+`</code> uploader is able to upload maven artifacts (whole GAV only!)
as artifact archive according to the maven artifact spec.
If registered the default mime type is: `+mime.MIME_TGZ+`

It accepts a plain string for the URL or a config with the following field:
'url': the URL of the maven repository.
`,
	)
}
