package index

import (
	"encoding/json"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
)

const SchemaVersion = 1

type ArtifactIndex struct {
	specs.Versioned
	Index []ArtifactMeta `json:"artifacts"`
}

type ArtifactMeta struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag,omitempty"`
	Digest     digest.Digest `json:"digest,omitempty"`
	MediaType  string        `json:"mediaType,omitempty"`
}

func Decode(data []byte) (*ArtifactIndex, error) {
	var d ArtifactIndex

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func Encode(d *ArtifactIndex) ([]byte, error) {
	return json.Marshal(d)
}
