// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/open-component-model/ocm/v2/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/accessmethods/options"
)

var (
	HintOption      = options.HintOption
	MediaTypeOption = options.MediatypeOption
)

var PathOption = flagsets.NewStringOptionType("inputPath", "path field for input")

var (
	CompressOption = flagsets.NewBoolOptionType("inputCompress", "compress option for input")
	ExcludeOption  = flagsets.NewStringArrayOptionType("inputExcludes", "excludes (path) for inputs")
)

var (
	IncludeOption     = flagsets.NewStringArrayOptionType("inputIncludes", "includes (path) for inputs")
	PreserveDirOption = flagsets.NewBoolOptionType("inputPreserveDir", "preserve directory in archive for inputs")
)

var (
	FollowSymlinksOption = flagsets.NewBoolOptionType("inputFollowSymlinks", "follow symbolic links during archive creation for inputs")
	VariantsOption       = flagsets.NewStringArrayOptionType("inputVariants", "(platform) variants for inputs")
)

var LibrariesOption = flagsets.NewStringArrayOptionType("inputLibraries", "library path for inputs")

var VersionOption = flagsets.NewStringArrayOptionType("inputVersion", "version info for inputs")

var ValuesOption = flagsets.NewValueMapYAMLOptionType("inputValues", "YAML based generic values for inputs")

var DataOption = flagsets.NewBytesOptionType("inputData", "data (string, !!string or !<base64>")

var TextOption = flagsets.NewStringOptionType("inputText", "utf8 text")
