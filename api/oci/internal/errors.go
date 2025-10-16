package internal

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

const (
	KIND_OCIARTIFACT = "oci artifact"
	KIND_BLOB        = blobaccess.KIND_BLOB
	KIND_MEDIATYPE   = blobaccess.KIND_MEDIATYPE
)

func ErrUnknownArtifact(name, version string) error {
	return errors.ErrUnknown(KIND_OCIARTIFACT, fmt.Sprintf("%s:%s", name, version))
}
