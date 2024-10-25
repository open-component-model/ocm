package ocireg

import (
	"encoding/json"
	"fmt"
	"path"

	containererr "github.com/containerd/containerd/remotes/errors"
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/extensions/pubsub/types/compound"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

const (
	ConfigMimeType     = "application/vnd.ocm.software.repository.config.v1+json"
	PubSubLayerMimeTye = "application/vnd.ocm.software.repository.config.pubsub.v1+json"
)

const META = "meta"

func init() {
	pubsub.RegisterProvider(ocireg.Type, &Provider{})
}

type Provider struct{}

var _ pubsub.Provider = (*Provider)(nil)

func (p *Provider) GetPubSubSpec(repo repocpi.Repository) (pubsub.PubSubSpec, error) {
	impl, err := repocpi.GetRepositoryImplementation(repo)
	if err != nil {
		return nil, err
	}
	gen, ok := impl.(*genericocireg.RepositoryImpl)
	if !ok {
		return nil, errors.ErrNotSupported("non-oci based ocm repository")
	}

	ocirepo := path.Join(gen.Meta().SubPath, componentmapping.ComponentDescriptorNamespace)
	acc, err := gen.OCIRepository().LookupArtifact(ocirepo, META)
	if errors.IsErrNotFound(err) || errors.IsErrUnknown(err) || errors.IsA(err, containererr.ErrUnexpectedStatus{}) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "cannot access meta data manifest version")
	}
	defer acc.Close()
	m := acc.ManifestAccess()
	if m == nil {
		return nil, fmt.Errorf("meta data artifact is no manifest artifact")
	}
	if m.GetDescriptor().Config.MediaType != ConfigMimeType {
		return nil, fmt.Errorf("meta data artifact has unexpected mime type %q", m.GetDescriptor().Config.MediaType)
	}
	compound, _ := compound.New()
	for _, l := range m.GetDescriptor().Layers {
		if l.MediaType == PubSubLayerMimeTye {
			var ps pubsub.GenericPubSubSpec

			blob, err := m.GetBlob(l.Digest)
			if err != nil {
				return nil, err
			}
			data, err := blob.Get()
			blob.Close()
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(data, &ps)
			if err != nil {
				return nil, err
			}
			compound.Specifications = append(compound.Specifications, &ps)
		}
	}
	return compound.Effective(), nil
}

func (p *Provider) SetPubSubSpec(repo cpi.Repository, spec pubsub.PubSubSpec) error {
	impl, err := repocpi.GetRepositoryImplementation(repo)
	if err != nil {
		return err
	}
	gen, ok := impl.(*genericocireg.RepositoryImpl)
	if !ok {
		return errors.ErrNotSupported("non-oci based ocm repository")
	}

	var data []byte
	if spec != nil {
		data, err = json.Marshal(spec)
		if err != nil {
			return err
		}
	}

	ocirepo := path.Join(gen.Meta().SubPath, componentmapping.ComponentDescriptorNamespace)
	ns, err := gen.OCIRepository().LookupNamespace(ocirepo)
	if err != nil {
		return err
	}
	defer ns.Close()

	acc, err := ns.GetArtifact(META)
	if err != nil {
		if errors.IsErrNotFound(err) || errors.IsErrUnknown(err) {
			if spec == nil {
				return nil
			}
		} else {
			return err
		}
	}
	if acc == nil {
		acc, err = ns.NewArtifact()
		if err != nil {
			return err
		}
		m, err := acc.Manifest()
		if err != nil {
			return err
		}
		config := blobaccess.ForString(ConfigMimeType, "{}")
		m.Config.MediaType = config.MimeType()
		m.Config.Digest = config.Digest()
		err = acc.AddBlob(config)
		if err != nil {
			return err
		}
	}
	defer acc.Close()

	m := acc.ManifestAccess()
	if m == nil {
		return fmt.Errorf("meta data artifact is no manifest artifact")
	}
	if m.GetDescriptor().Config.MediaType != ConfigMimeType {
		return fmt.Errorf("meta data artifact has unexpected mime type %q", m.GetDescriptor().Config.MediaType)
	}

	blob := blobaccess.ForData(PubSubLayerMimeTye, data)
	defer blob.Close()

	layers := m.GetDescriptor().Layers
	for i := 0; i < len(layers); i++ {
		l := layers[i]
		if l.MediaType == PubSubLayerMimeTye {
			if data != nil {
				m.AddBlob(blob)
				l.Digest = blob.Digest()
				b, err := ns.AddArtifact(m, META)
				if b != nil {
					b.Close()
				}
				return err
			} else {
				layers = append(layers[:i], layers[i+1:]...)
				i--
			}
		}
	}
	m.GetDescriptor().Layers = layers
	if data != nil {
		_, err = m.AddLayer(blob, nil)
		if err != nil {
			return err
		}
	}
	b, err := ns.AddArtifact(m, META)
	if b != nil {
		b.Close()
	}
	return err
}
