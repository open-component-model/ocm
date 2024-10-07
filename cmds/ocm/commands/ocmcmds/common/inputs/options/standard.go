package options

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var (
	HintOption      = RegisterOption(options.HintOption)
	MediatypeOption = RegisterOption(options.MediatypeOption)

	URLOption          = RegisterOption(options.URLOption)
	HTTPHeaderOption   = RegisterOption(options.HTTPHeaderOption)
	HTTPVerbOption     = RegisterOption(options.HTTPVerbOption)
	HTTPBodyOption     = RegisterOption(options.HTTPBodyOption)
	HTTPRedirectOption = RegisterOption(options.HTTPRedirectOption)

	GroupOption      = RegisterOption(options.GroupOption)
	ArtifactOption   = RegisterOption(options.ArtifactOption)
	ClassifierOption = RegisterOption(options.ClassifierOption)
	ExtensionOption  = RegisterOption(options.ExtensionOption)

	RegistryOption       = RegisterOption(options.NPMRegistryOption)
	PackageOption        = RegisterOption(options.NPMPackageOption)
	PackageVersionOption = RegisterOption(options.NPMVersionOption)

	IdentityPathOption = RegisterOption(options.IdentityPathOption)
)

// string options.
var (
	VersionOption        = RegisterOption(flagsets.NewStringOptionType("inputVersion", "version info for inputs"))
	TextOption           = RegisterOption(flagsets.NewStringOptionType("inputText", "utf8 text"))
	HelmRepositoryOption = RegisterOption(flagsets.NewStringOptionType("inputHelmRepository", "helm repository base URL"))
)

var (
	VariantsOption  = RegisterOption(flagsets.NewStringArrayOptionType("inputVariants", "(platform) variants for inputs"))
	PlatformsOption = RegisterOption(flagsets.NewStringArrayOptionType("inputPlatforms", "input filter for image platforms ([os]/[architecture])"))
)

// path options.
var (
	PathOption = RegisterOption(flagsets.NewStringOptionType("inputPath", "path field for input"))
)

var (
	IncludeOption   = RegisterOption(flagsets.NewStringArrayOptionType("inputIncludes", "includes (path) for inputs"))
	ExcludeOption   = RegisterOption(flagsets.NewStringArrayOptionType("inputExcludes", "excludes (path) for inputs"))
	LibrariesOption = RegisterOption(flagsets.NewStringArrayOptionType("inputLibraries", "library path for inputs"))
)

// boolean options.
var (
	CompressOption       = RegisterOption(flagsets.NewBoolOptionType("inputCompress", "compress option for input"))
	PreserveDirOption    = RegisterOption(flagsets.NewBoolOptionType("inputPreserveDir", "preserve directory in archive for inputs"))
	FollowSymlinksOption = RegisterOption(flagsets.NewBoolOptionType("inputFollowSymlinks", "follow symbolic links during archive creation for inputs"))
)

// data options.
var (
	DataOption = RegisterOption(flagsets.NewBytesOptionType("inputData", "data (string, !!string or !<base64>"))
)

// yaml/json options.
var (
	YAMLOption          = RegisterOption(flagsets.NewYAMLOptionType("inputYaml", "YAML formatted text"))
	JSONOption          = RegisterOption(flagsets.NewYAMLOptionType("inputJson", "JSON formatted text"))
	FormattedJSONOption = RegisterOption(flagsets.NewYAMLOptionType("inputFormattedJson", "JSON formatted text"))
)

var (
	ValuesOption    = RegisterOption(flagsets.NewValueMapYAMLOptionType("inputValues", "YAML based generic values for inputs"))
	ComponentOption = RegisterOption(flagsets.NewStringOptionType("inputComponent", "component name"))
)

// RepositoryOption sets the repository or registry for an input.
var RepositoryOption = RegisterOption(flagsets.NewStringOptionType("inputRepository", "repository or registry for inputs"))
