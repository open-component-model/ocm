package grammar

import (
	. "github.com/mandelsoft/goutils/regexutils"

	"ocm.software/ocm/api/oci/grammar"
)

const (
	ComponentSeparatorChar = grammar.RepositorySeparatorChar
	ComponentSeparator     = grammar.RepositorySeparator

	VersionSeparatorChar = grammar.TagSeparatorChar
	VersionSeparator     = grammar.TagSeparator
)

var (
	// TypeRegexp describes a type name for a repository.
	TypeRegexp = grammar.TypeRegexp

	VersionRegexp = Sequence(Optional(Literal("v")), Numeric, Repetition(0, 2, Literal("."), Numeric),
		Optional(Literal("-"), Repeated(CharSet("0-9A-Za-z-")), OptionalRepeated(Literal("."), Repeated(CharSet("0-9A-Za-z-")))),
		Optional(Literal("+"), Repeated(CharSet("0-9A-Za-z-"))), OptionalRepeated(Literal("."), Repeated(CharSet("0-9A-Za-z-"))),
	)

	// AnchoredRepositoryRegexp parses a uniform repository spec.
	AnchoredRepositoryRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		grammar.SchemeDomainPortRegexp, Optional(grammar.RepositorySeparatorRegexp, Capture(grammar.RepositoryRegexp)),
	)

	// AnchoredSchemedHostPortRepositoryRegexp parses a uniform repository spec.
	AnchoredSchemedHostPortRepositoryRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		grammar.SchemedHostPortRegexp, Optional(grammar.RepositorySeparatorRegexp, Capture(grammar.RepositoryRegexp)),
	)

	// AnchoredHostWithPortRepositoryRegexp parses a uniform repository spec.
	AnchoredHostWithPortRepositoryRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		grammar.SchemeHostPortRegexp, Optional(grammar.RepositorySeparatorRegexp, Capture(grammar.RepositoryRegexp)),
	)

	// AnchoredGenericRepositoryRegexp describes a CTF reference.
	AnchoredGenericRepositoryRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		Capture(Match(".*")),
	)

	// ComponentRegexp describes the component name. It consists
	// of a domain name followed by OCI repository name components.
	ComponentRegexp = Sequence(grammar.DomainRegexp, grammar.RepositorySeparatorRegexp, grammar.RepositoryRegexp)

	// AnchoredComponentVersionRegexp parses a component with an optional version.
	AnchoredComponentVersionRegexp = Anchored(
		Capture(ComponentRegexp),
		Optional(Literal(VersionSeparator), Capture(VersionRegexp)),
	)

	// AnchoredReferenceRegexp parses a complete string representation for default component references including
	// the repository part.
	// It provides 5 captures: type, repository host port, sub path, component and version.
	AnchoredReferenceRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		grammar.SchemeDomainPortRegexp, Optional(grammar.RepositorySeparatorRegexp, Capture(grammar.RepositoryRegexp)),
		Literal("//"), Capture(ComponentRegexp),
		Optional(Literal(VersionSeparator), Capture(VersionRegexp)),
	)

	// AnchoredSchemedHostPortReferenceRegexp parses a complete string representation for default component references
	// including the repository part. Since the type is optional, the scheme is required to allow for a distinction
	// from filepaths.
	// It provides 6 captures: type, scheme, repository host port, sub path, component and version.
	AnchoredSchemedHostPortReferenceRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		grammar.SchemedHostPortRegexp, Optional(grammar.RepositorySeparatorRegexp, Capture(grammar.RepositoryRegexp)),
		Literal("//"), Capture(ComponentRegexp),
		Optional(Literal(VersionSeparator), Capture(VersionRegexp)),
	)

	// AnchoredHostWithPortReferenceRegexp parses a complete string representation for default component references
	// including the repository part. Since the type is optional, the scheme is required to allow for a distinction
	// from filepaths.
	// It provides 6 captures: type, scheme, repository host port, sub path, component and version.
	AnchoredHostWithPortReferenceRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		grammar.SchemeHostPortRegexp, Optional(grammar.RepositorySeparatorRegexp, Capture(grammar.RepositoryRegexp)),
		Literal("//"), Capture(ComponentRegexp),
		Optional(Literal(VersionSeparator), Capture(VersionRegexp)),
	)

	// AnchoredGenericReferenceRegexp parses a CTF file based string representation.
	AnchoredGenericReferenceRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		Capture(Match(".*?")),
		Optional(Literal("//"), Capture(ComponentRegexp),
			Optional(Literal(VersionSeparator), Capture(VersionRegexp))),
	)
)

func SplitTypeSpec(t string) (string, string) {
	return grammar.SplitTypeSpec(t)
}
