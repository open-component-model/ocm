package internal

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/utils/errkind"
)

const (
	KIND_REPOSITORY     = "ocm repository"
	KIND_REPOSITORYSPEC = "repository specification"

	KIND_COMPONENT        = errkind.KIND_COMPONENT
	KIND_COMPONENTVERSION = compdesc.KIND_COMPONENTVERSION
	KIND_RESOURCE         = compdesc.KIND_RESOURCE
	KIND_SOURCE           = compdesc.KIND_SOURCE
	KIND_REFERENCE        = compdesc.KIND_REFERENCE
	KIND_OCM_REFERENCE    = "ocm reference"
)

func ErrComponentVersionNotFound(name, version string) error {
	return errors.ErrNotFound(KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", name, version))
}

func ErrComponentVersionNotFoundWrap(err error, name, version string) error {
	return errors.ErrNotFoundWrap(err, KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", name, version))
}
