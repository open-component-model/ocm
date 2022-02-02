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

package artefactset

import (
	"reflect"

	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type StateHandler struct {
	fs vfs.FileSystem
}

var _ accessobj.StateHandler = &StateHandler{}

func NewStateHandler(fs vfs.FileSystem) accessobj.StateHandler {
	return &StateHandler{fs}
}

func (i StateHandler) Initial() interface{} {
	return artdesc.NewIndex()
}

func (i StateHandler) Encode(d interface{}) ([]byte, error) {
	return artdesc.EncodeIndex(d.(*artdesc.Index))
}

func (i StateHandler) Decode(data []byte) (interface{}, error) {
	return artdesc.DecodeIndex(data)
}

func (i StateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
