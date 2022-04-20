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
	"github.com/open-component-model/ocm/pkg/errors"
)

const KIND_CREDENTIALS = "credentials"
const KIND_CONSUMER = "consumer"
const KIND_REPOSITORY = "repository"

func ErrUnknownCredentials(name string) error {
	return errors.ErrUnknown(KIND_CREDENTIALS, name)
}

func ErrUnknownConsumer(name string) error {
	return errors.ErrUnknown(KIND_CONSUMER, name)
}

func ErrUnknownRepository(kind, name string) error {
	return errors.ErrUnknown(KIND_REPOSITORY, name, kind)
}
