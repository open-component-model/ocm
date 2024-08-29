package options

// HintOption .
var HintOption = RegisterOption(NewStringOptionType("hint", "(repository) hint for local artifacts"))

// MediatypeOption .
var MediatypeOption = RegisterOption(NewStringOptionType("mediaType", "media type for artifact blob representation"))

// SizeOption .
var SizeOption = RegisterOption(NewIntOptionType("size", "blob size"))

// DigestOption .
var DigestOption = RegisterOption(NewStringOptionType("digest", "blob digest"))

// ReferenceOption .
var ReferenceOption = RegisterOption(NewStringOptionType("reference", "reference name"))

// PackageOption .
var PackageOption = RegisterOption(NewStringOptionType("package", "package or object name"))

// ArtifactOption .
var ArtifactOption = RegisterOption(NewStringOptionType("artifactId", "maven artifact id"))

// GroupOption .
var GroupOption = RegisterOption(NewStringOptionType("groupId", "maven group id"))

// RepositoryOption .
var RepositoryOption = RegisterOption(NewStringOptionType("accessRepository", "repository URL"))

// RegistryOption .
var RegistryOption = RegisterOption(NewStringOptionType("accessRegistry", "registry base URL"))

// HostnameOption .
var HostnameOption = RegisterOption(NewStringOptionType("accessHostname", "hostname used for access"))

// CommitOption .
var CommitOption = RegisterOption(NewStringOptionType("commit", "git commit id"))

// GlobalAccessOption .
var GlobalAccessOption = RegisterOption(NewValueMapYAMLOptionType("globalAccess", "access specification for global access"))

// RegionOption .
var RegionOption = RegisterOption(NewStringOptionType("region", "region name"))

// BucketOption .
var BucketOption = RegisterOption(NewStringOptionType("bucket", "bucket name"))

// VersionOption .
var VersionOption = RegisterOption(NewStringOptionType("accessVersion", "version for access specification"))

// ComponentOption.
var ComponentOption = RegisterOption(NewStringOptionType("accessComponent", "component for access specification"))

// IdentityPathOption.
var IdentityPathOption = RegisterOption(NewIdentityPathOptionType("identityPath", "identity path for specification"))

// URLOption.
var URLOption = RegisterOption(NewStringOptionType("url", "artifact or server url"))

var HTTPHeaderOption = RegisterOption(NewStringSliceMapColonOptionType("header", "http headers"))

var HTTPVerbOption = RegisterOption(NewStringOptionType("verb", "http request method"))

var HTTPBodyOption = RegisterOption(NewStringOptionType("body", "body of a http request"))

var HTTPRedirectOption = RegisterOption(NewBoolOptionType("noredirect", "http redirect behavior"))

// CommentOption .
var CommentOption = RegisterOption(NewStringOptionType("comment", "comment field value"))

// ClassifierOption the optional classifier of a maven resource.
var ClassifierOption = RegisterOption(NewStringOptionType("classifier", "maven classifier"))

// ExtensionOption the optional extension of a maven resource.
var ExtensionOption = RegisterOption(NewStringOptionType("extension", "maven extension name"))

// NPMRegistryOption sets the registry of the npm resource.
var NPMRegistryOption = RegisterOption(NewStringOptionType("registry", "npm package registry"))

// NPMPackageOption sets what package should be fetched from the npm registry.
var NPMPackageOption = RegisterOption(NewStringOptionType("package", "npm package name"))

// NPMVersionOption sets the version of the npm package.
var NPMVersionOption = RegisterOption(NewStringOptionType("version", "npm package version"))

// IdPathOption is a path of identity specs.
var IdPathOption = RegisterOption(NewStringArrayOptionType("idpath", "identity path (attr=value{,attr=value}"))
