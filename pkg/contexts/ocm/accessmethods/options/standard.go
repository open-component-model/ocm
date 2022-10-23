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

package options

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

// HintOption.
var HintOption = flagsets.NewStringOptionType("hint", "(repository) hint for local artifacts")

// MediatypeOption.
var MediatypeOption = flagsets.NewStringOptionType("mediaType", "media type for artifact blob representation")

// SizeOption.
var SizeOption = flagsets.NewIntOptionType("size", "blob size")

// DigestOption.
var DigestOption = flagsets.NewStringOptionType("digest", "blob digest")

// ReferenceOption.
var ReferenceOption = flagsets.NewStringOptionType("reference", "reference name")

// RepositoryOption.
var RepositoryOption = flagsets.NewStringOptionType("accessRepository", "repository URL")

// HostnameOption.
var HostnameOption = flagsets.NewStringOptionType("accessHostname", "hostname used for access")

// CommitOption.
var CommitOption = flagsets.NewStringOptionType("commit", "git commit id")

// GlobalAccessOption.
var GlobalAccessOption = flagsets.NewValueMapOptionType("globalAccess", "access specification for global access")

// RegionOption.
var RegionOption = flagsets.NewStringOptionType("region", "region name")

// BucketOption.
var BucketOption = flagsets.NewStringOptionType("bucket", "bucket name")

// VersionOption.
var VersionOption = flagsets.NewStringOptionType("accessVersion", "version for access specification")
