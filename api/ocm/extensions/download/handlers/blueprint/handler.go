package blueprint

import (
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	registry "ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	TYPE             = resourcetypes.BLUEPRINT
	LEGACY_TYPE      = resourcetypes.BLUEPRINT_LEGACY
	CONFIG_MIME_TYPE = "application/vnd.gardener.landscaper.blueprint.config.v1"
)

type Extractor func(pr common.Printer, handler *Handler, access blobaccess.DataAccess, path string, fs vfs.FileSystem) (bool, error)

var (
	supportedArtifactTypes    []string
	mimeTypeExtractorRegistry map[string]Extractor
)

type Handler struct {
	ociConfigMimeTypes set.Set[string]
}

func init() {
	supportedArtifactTypes = []string{TYPE, LEGACY_TYPE}
	mimeTypeExtractorRegistry = map[string]Extractor{
		mime.MIME_TAR:                        ExtractArchive,
		mime.MIME_TGZ:                        ExtractArchive,
		mime.MIME_TGZ_ALT:                    ExtractArchive,
		BLUEPRINT_MIMETYPE:                   ExtractArchive,
		BLUEPRINT_MIMETYPE_COMPRESSED:        ExtractArchive,
		BLUEPRINT_MIMETYPE_LEGACY:            ExtractArchive,
		BLUEPRINT_MIMETYPE_LEGACY_COMPRESSED: ExtractArchive,
	}
	for _, t := range append(artdesc.ToArchiveMediaTypes(artdesc.MediaTypeImageManifest), artdesc.ToArchiveMediaTypes(artdesc.MediaTypeDockerSchema2Manifest)...) {
		mimeTypeExtractorRegistry[t] = ExtractArtifact
	}

	h := New()

	registry.Register(h, registry.ForArtifactType(TYPE))
	registry.Register(h, registry.ForArtifactType(LEGACY_TYPE))
}

func New(configmimetypes ...string) *Handler {
	if len(configmimetypes) == 0 || utils.Optional(configmimetypes...) == "" {
		configmimetypes = []string{CONFIG_MIME_TYPE}
	}
	return &Handler{
		ociConfigMimeTypes: set.New[string](configmimetypes...),
	}
}

func (h *Handler) Download(pr common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (_ bool, _ string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagationf(&err, "downloading blueprint")

	meth, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	finalize.Close(meth)

	ex := mimeTypeExtractorRegistry[meth.MimeType()]
	if ex == nil {
		return false, "", nil
	}

	ok, err := ex(pr, h, meth, path, fs)
	if err != nil || !ok {
		return ok, "", err
	}
	return true, path, nil
}
