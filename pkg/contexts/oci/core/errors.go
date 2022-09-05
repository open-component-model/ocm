// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
