package artdesc

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type Manifest ociv1.Manifest

var _ BlobDescriptorSource = (*Manifest)(nil)

func NewManifest() *Manifest {
	return &Manifest{
		Versioned:   specs.Versioned{SchemeVersion},
		MediaType:   MediaTypeImageManifest,
		Layers:      nil,
		Annotations: nil,
	}
}

var _ ArtifactDescriptor = (*Manifest)(nil)

func (i *Manifest) IsManifest() bool {
	return true
}

func (i *Manifest) IsIndex() bool {
	return false
}

func (i *Manifest) Digest() digest.Digest {
	blob, _ := i.Blob()
	if blob != nil {
		return blob.Digest()
	}
	return ""
}

func (i *Manifest) Artifact() *Artifact {
	return &Artifact{manifest: i}
}

func (i *Manifest) Manifest() (*Manifest, error) {
	return i, nil
}

func (i *Manifest) Index() (*Index, error) {
	return nil, errors.ErrInvalid()
}

func (i *Manifest) IsValid() bool {
	return true
}

func (m *Manifest) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	if m.Config.Digest == digest {
		d := m.Config
		return &d
	}
	for _, l := range m.Layers {
		if l.Digest == digest {
			return &l
		}
	}
	return nil
}

func (m *Manifest) MimeType() string {
	return ArtifactMimeType(m.MediaType, MediaTypeImageManifest, legacy)
}

func (m *Manifest) Blob() (blobaccess.BlobAccess, error) {
	m.MediaType = m.MimeType()
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return blobaccess.ForData(m.MediaType, data), nil
}

func (m *Manifest) SetAnnotation(name, value string) {
	if m.Annotations == nil {
		m.Annotations = map[string]string{}
	}
	m.Annotations[name] = value
}

func (m *Manifest) DeleteAnnotation(name string) {
	if m.Annotations == nil {
		return
	}
	delete(m.Annotations, name)
	if len(m.Annotations) == 0 {
		m.Annotations = nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func DecodeManifest(data []byte) (*Manifest, error) {
	var d Manifest

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func EncodeManifest(d *Manifest) ([]byte, error) {
	return json.Marshal(d)
}
