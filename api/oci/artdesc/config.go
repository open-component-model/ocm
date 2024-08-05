package artdesc

import (
	"encoding/json"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type ImageConfig = ociv1.Image

func ParseImageConfig(blob blobaccess.BlobAccess) (*ImageConfig, error) {
	var cfg ImageConfig

	data, err := blob.Get()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
