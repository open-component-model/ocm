package options

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

// HintOption .
var HintOption = RegisterOption(flagsets.NewStringOptionType("hint", "(repository) hint for local artifacts"))

// MediatypeOption .
var MediatypeOption = RegisterOption(flagsets.NewStringOptionType("mediaType", "media type for artifact blob representation"))

// SizeOption .
var SizeOption = RegisterOption(flagsets.NewIntOptionType("size", "blob size"))

// DigestOption .
var DigestOption = RegisterOption(flagsets.NewStringOptionType("digest", "blob digest"))

// ReferenceOption .
var ReferenceOption = RegisterOption(flagsets.NewStringOptionType("reference", "reference name"))

// PackageOption .
var PackageOption = RegisterOption(flagsets.NewStringOptionType("package", "package or object name"))

// ArtifactOption .
var ArtifactOption = RegisterOption(flagsets.NewStringOptionType("artifactId", "maven artifact id"))

// GroupOption .
var GroupOption = RegisterOption(flagsets.NewStringOptionType("groupId", "maven group id"))

// RepositoryOption .
var RepositoryOption = RegisterOption(flagsets.NewStringOptionType("accessRepository", "repository or registry URL"))

// HostnameOption .
var HostnameOption = RegisterOption(flagsets.NewStringOptionType("accessHostname", "hostname used for access"))

// CommitOption .
var CommitOption = RegisterOption(flagsets.NewStringOptionType("commit", "git commit id"))

// GlobalAccessOption .
var GlobalAccessOption = RegisterOption(flagsets.NewValueMapYAMLOptionType("globalAccess", "access specification for global access"))

// RegionOption .
var RegionOption = RegisterOption(flagsets.NewStringOptionType("region", "region name"))

// BucketOption .
var BucketOption = RegisterOption(flagsets.NewStringOptionType("bucket", "bucket name"))

// VersionOption .
var VersionOption = RegisterOption(flagsets.NewStringOptionType("accessVersion", "version for access specification"))

// ComponentOption.
var ComponentOption = RegisterOption(flagsets.NewStringOptionType("accessComponent", "component for access specification"))

// IdentityPathOption.
var IdentityPathOption = RegisterOption(flagsets.NewIdentityPathOptionType("identityPath", "identity path for specification"))

// URLOption.
var URLOption = RegisterOption(flagsets.NewStringOptionType("url", "artifact or server url"))

var HTTPHeaderOption = RegisterOption(flagsets.NewStringSliceMapColonOptionType("header", "http headers"))

var HTTPVerbOption = RegisterOption(flagsets.NewStringOptionType("verb", "http request method"))

var HTTPBodyOption = RegisterOption(flagsets.NewStringOptionType("body", "body of a http request"))

var HTTPRedirectOption = RegisterOption(flagsets.NewBoolOptionType("noredirect", "http redirect behavior"))

// CommentOption .
var CommentOption = RegisterOption(flagsets.NewStringOptionType("comment", "comment field value"))

// ClassifierOption the optional classifier of a maven resource.
var ClassifierOption = RegisterOption(flagsets.NewStringOptionType("classifier", "maven classifier"))

// ExtensionOption the optional extension of a maven resource.
var ExtensionOption = RegisterOption(flagsets.NewStringOptionType("extension", "maven extension name"))

// NPMRegistryOption sets the registry of the npm resource.
var NPMRegistryOption = RegisterOption(flagsets.NewStringOptionType("registry", "npm package registry"))

// NPMPackageOption sets what package should be fetched from the npm registry.
var NPMPackageOption = PackageOption

// NPMVersionOption sets the version of the npm package.
var NPMVersionOption = RegisterOption(flagsets.NewStringOptionType("version", "npm package version"))

// IdPathOption is a path of identity specs.
var IdPathOption = RegisterOption(flagsets.NewStringArrayOptionType("idpath", "identity path (attr=value{,attr=value}"))
