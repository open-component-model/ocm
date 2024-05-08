package internal

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	KIND_COMPONENT        = errors.KIND_COMPONENT
	KIND_COMPONENTVERSION = "component version"
	KIND_RESOURCE         = "component resource"
	KIND_SOURCE           = "component source"
	KIND_REFERENCE        = compdesc.KIND_REFERENCE
	KIND_REPOSITORYSPEC   = "repository specification"
)

func ErrComponentVersionNotFound(name, version string) error {
	return errors.ErrNotFound(KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", name, version))
}

func ErrComponentVersionNotFoundWrap(err error, name, version string) error {
	return errors.ErrNotFoundWrap(err, KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", name, version))
}
