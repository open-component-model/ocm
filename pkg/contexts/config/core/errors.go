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

	"github.com/open-component-model/ocm/pkg/errors"
)

const KIND_CONFIGTYPE = "config type"

// OCM_CONFIG_SUFFIX is the standard suffix used for all configuration types
// provided by this library.
const OCM_CONFIG_SUFFIX = ".config.ocm.software"

////////////////////////////////////////////////////////////////////////////////

type noContextError struct {
	name string
}

func (e *noContextError) Error() string {
	return fmt.Sprintf("unknown context %q", e.name)
}

func ErrNoContext(name string) error {
	return &noContextError{name}
}

func IsErrNoContext(err error) bool {
	return errors.IsA(err, &noContextError{})
}

func IsErrConfigNotApplicable(err error) bool {
	return errors.IsErrUnknownKind(err, KIND_CONFIGTYPE) || IsErrNoContext(err)
}
