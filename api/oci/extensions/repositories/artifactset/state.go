package artifactset

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/api/oci/cpi"
	"github.com/open-component-model/ocm/api/utils/accessobj"
)

// NewStateHandler implements the factory interface for the artifact set
// state descriptor handling
// Basically this is an index state.
func NewStateHandler(fs vfs.FileSystem) accessobj.StateHandler {
	return &cpi.IndexStateHandler{}
}
