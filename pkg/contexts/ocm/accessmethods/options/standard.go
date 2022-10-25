// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
var GlobalAccessOption = flagsets.NewValueMapYAMLOptionType("globalAccess", "access specification for global access")

// RegionOption.
var RegionOption = flagsets.NewStringOptionType("region", "region name")

// BucketOption.
var BucketOption = flagsets.NewStringOptionType("bucket", "bucket name")

// VersionOption.
var VersionOption = flagsets.NewStringOptionType("accessVersion", "version for access specification")
