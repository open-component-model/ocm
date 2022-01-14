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

package accessmethods

import (
	"github.com/gardener/ocm/pkg/ocm/common"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf"
)

// LocalFilesystemBlobType is the access type of a blob in a local filesystem.
const LocalFilesystemBlobType = ctf.LocalFilesystemBlobType
const LocalFilesystemBlobTypeV1 = ctf.LocalFilesystemBlobTypeV1

func init() {
	core.RegisterAccessType(common.NewAccessType(LocalFilesystemBlobType, &LocalFilesystemBlobAccessSpec{}))
	core.RegisterAccessType(common.NewAccessType(LocalFilesystemBlobTypeV1, &LocalFilesystemBlobAccessSpec{}))
}

// NewLocalFilesystemBlobAccessSpecV1 creates a new localFilesystemBlob accessor.
var NewLocalFilesystemBlobAccessSpecV1 = ctf.NewLocalFilesystemBlobAccessSpecV1

// LocalFilesystemBlobAccessSpec describes the access for a blob on the filesystem.
type LocalFilesystemBlobAccessSpec = ctf.LocalFilesystemBlobAccessSpec
