// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

// HintOption.
var HintOption = RegisterOption(flagsets.NewStringOptionType("hint", "(repository) hint for local artifacts"))

// MediatypeOption.
var MediatypeOption = RegisterOption(flagsets.NewStringOptionType("mediaType", "media type for artifact blob representation"))

// SizeOption.
var SizeOption = RegisterOption(flagsets.NewIntOptionType("size", "blob size"))

// DigestOption.
var DigestOption = RegisterOption(flagsets.NewStringOptionType("digest", "blob digest"))

// ReferenceOption.
var ReferenceOption = RegisterOption(flagsets.NewStringOptionType("reference", "reference name"))

// RepositoryOption.
var RepositoryOption = RegisterOption(flagsets.NewStringOptionType("accessRepository", "repository URL"))

// HostnameOption.
var HostnameOption = RegisterOption(flagsets.NewStringOptionType("accessHostname", "hostname used for access"))

// CommitOption.
var CommitOption = RegisterOption(flagsets.NewStringOptionType("commit", "git commit id"))

// GlobalAccessOption.
var GlobalAccessOption = RegisterOption(flagsets.NewValueMapYAMLOptionType("globalAccess", "access specification for global access"))

// RegionOption.
var RegionOption = RegisterOption(flagsets.NewStringOptionType("region", "region name"))

// BucketOption.
var BucketOption = RegisterOption(flagsets.NewStringOptionType("bucket", "bucket name"))

// VersionOption.
var VersionOption = flagsets.NewStringOptionType("accessVersion", "version for access specification")
