package options

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

var (
	HintOption      = options.HintOption
	MediaTypeOption = options.MediatypeOption
)

// string options
var (
	VersionOption        = flagsets.NewStringOptionType("inputVersion", "version info for inputs")
	TextOption           = flagsets.NewStringOptionType("inputText", "utf8 text")
	HelmRepositoryOption = flagsets.NewStringOptionType("inputHelmRepository", "helm repository base URL")
)
var (
	VariantsOption  = flagsets.NewStringArrayOptionType("inputVariants", "(platform) variants for inputs")
	PlatformsOption = flagsets.NewStringArrayOptionType("inputPlatforms", "input filter for image platforms ([os]/[architecture])")
)

// path options
var (
	PathOption = flagsets.NewPathOptionType("inputPath", "path field for input")
)
var (
	IncludeOption   = flagsets.NewPathArrayOptionType("inputIncludes", "includes (path) for inputs")
	ExcludeOption   = flagsets.NewPathArrayOptionType("inputExcludes", "excludes (path) for inputs")
	LibrariesOption = flagsets.NewPathArrayOptionType("inputLibraries", "library path for inputs")
)

// boolean options
var (
	CompressOption       = flagsets.NewBoolOptionType("inputCompress", "compress option for input")
	PreserveDirOption    = flagsets.NewBoolOptionType("inputPreserveDir", "preserve directory in archive for inputs")
	FollowSymlinksOption = flagsets.NewBoolOptionType("inputFollowSymlinks", "follow symbolic links during archive creation for inputs")
)

// data options
var (
	DataOption = flagsets.NewBytesOptionType("inputData", "data (string, !!string or !<base64>")
)

// yaml/json options
var (
	YAMLOption          = flagsets.NewYAMLOptionType("inputYaml", "YAML formatted text")
	JSONOption          = flagsets.NewYAMLOptionType("inputJson", "JSON formatted text")
	FormattedJSONOption = flagsets.NewYAMLOptionType("inputFormattedJson", "JSON formatted text")
)
var (
	ValuesOption = flagsets.NewValueMapYAMLOptionType("inputValues", "YAML based generic values for inputs")
)
