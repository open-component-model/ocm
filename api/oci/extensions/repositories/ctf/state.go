package ctf

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci/extensions/repositories/ctf/index"
	"ocm.software/ocm/api/utils/accessobj"
)

type StateHandler struct{}

var _ accessobj.StateHandler = &StateHandler{}

func NewStateHandler(fs vfs.FileSystem) accessobj.StateHandler {
	return &StateHandler{}
}

func (i StateHandler) Initial() interface{} {
	return index.NewRepositoryIndex()
}

func (i StateHandler) Encode(d interface{}) ([]byte, error) {
	return index.Encode(d.(*index.RepositoryIndex).GetDescriptor())
}

func (i StateHandler) Decode(data []byte) (interface{}, error) {
	idx, err := index.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse artifact index read from %s: %w", ArtifactIndexFileName, err)
	}
	if idx.SchemaVersion != index.SchemaVersion {
		return nil, fmt.Errorf("unknown schema version %d for artifact index %s", index.SchemaVersion, ArtifactIndexFileName)
	}

	artifacts := index.NewRepositoryIndex()
	for i := range idx.Index {
		artifacts.AddArtifactInfo(&idx.Index[i])
	}
	return artifacts, nil
}

func (i StateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
