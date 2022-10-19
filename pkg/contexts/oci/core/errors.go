// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	KIND_OCIARTEFACT = "oci artefact"
	KIND_BLOB        = accessio.KIND_BLOB
	KIND_MEDIATYPE   = accessio.KIND_MEDIATYPE
)

func ErrUnknownArtefact(name, version string) error {
	return errors.ErrUnknown(KIND_OCIARTEFACT, fmt.Sprintf("%s:%s", name, version))
}
