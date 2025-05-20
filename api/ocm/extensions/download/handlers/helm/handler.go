package helm

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/vfs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	helmregistry "helm.sh/helm/v3/pkg/registry"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	registry "ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
)

const TYPE = resourcetypes.HELM_CHART

var (
	// ErrNoMatchingLayer denotes that no layer matching the requested media type was found,,
	// which is invalid as per HELM OCI specification for [helmregistry.ChartLayerMediaType]
	// but valid for [helmregistry.ProvLayerMediaType].
	ErrNoMatchingLayer = errors.New("no matching layer found")
	// ErrMultipleMatchingLayers denotes that multiple layers matching the requested media type were found,
	// which is invalid as per HELM OCI specification.
	ErrMultipleMatchingLayers = errors.New("multiple matching layers found")
)

type Handler struct{}

func init() {
	registry.Register(New(), registry.ForArtifactType(TYPE))
}

func New() *Handler {
	return &Handler{}
}

func AssureArchiveSuffix(name string) string {
	if !strings.HasSuffix(name, ".tgz") && !strings.HasSuffix(name, ".tar.gz") {
		name += ".tgz"
	}
	return name
}

func (h Handler) fromArchive(p common.Printer, meth cpi.AccessMethod, path string, fs vfs.FileSystem) (_ bool, _ string, err error) {
	basetype := mime.BaseType(helmregistry.ChartLayerMediaType)
	if mime.BaseType(meth.MimeType()) != basetype {
		return false, "", nil
	}

	chart := AssureArchiveSuffix(path)

	err = write(p, meth, chart, fs)
	if err != nil {
		return true, "", err
	}
	return true, chart, nil
}

func (h Handler) fromOCIArtifact(p common.Printer, meth cpi.AccessMethod, path string, fs vfs.FileSystem) (_ bool, _ string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagationf(&err, "from OCI artifact")

	rd, err := meth.Reader()
	if err != nil {
		return true, "", err
	}
	finalize.Close(rd, "access method reader")
	set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(rd))
	if err != nil {
		return true, "", err
	}
	finalize.Close(set, "artifact set")
	art, err := set.GetArtifact(set.GetMain().String())
	if err != nil {
		return true, "", err
	}
	finalize.Close(art)
	chart, _, err := download(p, art, path, fs)
	if err != nil {
		return true, "", err
	}
	return true, chart, nil
}

func (h Handler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (_ bool, _ string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagationf(&err, "downloading helm chart")

	if path == "" {
		path = racc.Meta().GetName()
	}

	meth, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	finalize.Close(meth)
	if mime.BaseType(meth.MimeType()) != mime.BaseType(artdesc.MediaTypeImageManifest) {
		return h.fromArchive(p, meth, path, fs)
	}
	return h.fromOCIArtifact(p, meth, path, fs)
}

// download downloads the chart and optional provenance file from an  oci.ArtifactAccess.
// the format of the artifact is expected to match the official HELM Reference Specification
//
// see https://github.com/helm/community/blob/dd5fe7878e293c573cc22db5d36558709c7b8a43/hips/hip-0006.md
func download(p common.Printer, art oci.ArtifactAccess, path string, fs vfs.FileSystem) (chart, prov string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	m := art.ManifestAccess()
	if m == nil {
		return "", "", errors.New("artifact is no image manifest")
	}
	desc := m.GetDescriptor()

	chartDesc, err := findLayer(desc.Layers, helmregistry.ChartLayerMediaType)
	if err != nil {
		return "", "", fmt.Errorf("no valid chart layer found: %w", err)
	}

	chartBlob, err := m.GetBlob(chartDesc.Digest)
	if err != nil {
		return "", "", fmt.Errorf("no valid chart blob found: %w", err)
	}
	finalize.Close(chartBlob)

	chart = AssureArchiveSuffix(path)
	if err := write(p, chartBlob, chart, fs); err != nil {
		return "", "", err
	}

	// Optional provenance layer, if present, add it separately
	if provDesc, err := findLayer(desc.Layers, helmregistry.ProvLayerMediaType); err == nil {
		provBlob, err := m.GetBlob(provDesc.Digest)
		if err != nil {
			return "", "", err
		}
		finalize.Close(provBlob)

		prov = chart[:len(chart)-3] + "prov"
		if err := write(p, provBlob, path, fs); err != nil {
			return "", "", err
		}
	} else if !errors.Is(err, ErrNoMatchingLayer) { // Ignore if no provenance layer is found, because its optional.
		return "", "", err
	}

	return chart, prov, nil
}

func findLayer(layers []ocispec.Descriptor, mediaType string) (*ocispec.Descriptor, error) {
	var candidates []*ocispec.Descriptor

	for _, l := range layers {
		if mime.BaseType(l.MediaType) == mime.BaseType(mediaType) {
			candidates = append(candidates, &l)
		}
	}

	switch {
	case len(candidates) > 1:
		return nil, fmt.Errorf("%w: %s", ErrMultipleMatchingLayers, mediaType)
	case len(candidates) == 0:
		return nil, fmt.Errorf("%w: %s", ErrNoMatchingLayer, mediaType)
	default:
		return candidates[0], nil
	}
}

func write(p common.Printer, blob blobaccess.DataReader, path string, fs vfs.FileSystem) (err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	cr, err := blob.Reader()
	if err != nil {
		return err
	}
	finalize.Close(cr)
	file, err := fs.OpenFile(path, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0o660)
	if err != nil {
		return err
	}
	finalize.Close(file)
	n, err := io.Copy(file, cr)
	if err == nil {
		p.Printf("%s: %d byte(s) written\n", path, n)
	}
	return nil
}
