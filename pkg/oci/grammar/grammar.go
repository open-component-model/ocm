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

package grammar

import (
	. "github.com/gardener/ocm/pkg/regex"
)

const (
	// RepositorySeparatorChar is the separator character used to separate
	// repository name components.
	RepositorySeparatorChar = '/'

	// RepositorySeparator is the separator string used to separate
	// repository name components.
	RepositorySeparator = string(RepositorySeparatorChar)

	TagSeparatorChar = ":"
	TagSeparator     = string(TagSeparatorChar)

	DigestSeparatorChar = "@"
	DigestSeparator     = string(DigestSeparatorChar)
)

var (
	// TypeRegexp describes a type name for a repository
	TypeRegexp = Optional(Identifier)

	// CapturedSchemeRegexp matches an optional scheme
	CapturedSchemeRegexp = Sequence(Capture(Match(`[a-z]+`)), Match("://"))

	// AnchoredRegistryRegexp parses a uniform respository spec
	AnchoredRegistryRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		Optional(CapturedSchemeRegexp),
		Capture(DomainPortRegexp),
	)

	// AnchoredGenericRegistryRegexp describes a CTF reference
	AnchoredGenericRegistryRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		Capture(Match(".*")),
	)

	// RepositorySeparatorRegexp is the separator used to separate
	// repository name components.
	RepositorySeparatorRegexp = Literal(RepositorySeparator)

	// alphaNumericRegexp defines the alpha numeric atom, typically a
	// component of names. This only allows lower case characters and digits.
	AlphaNumericRegexp = Match(`[a-z0-9]+`)

	// separatorRegexp defines the separators allowed to be embedded in name
	// components. This allow one period, one or two underscore and multiple
	// dashes.
	separatorRegexp = Match(`(?:[._]|__|[-]*)`)

	// NameComponentRegexp restricts registry path component names to start
	// with at least one letter or number, with following parts able to be
	// separated by one period, one or two underscore and multiple dashes.
	NameComponentRegexp = Sequence(
		AlphaNumericRegexp,
		Optional(Repeated(separatorRegexp, AlphaNumericRegexp)))

	// DomainComponentRegexp restricts the registry domain component of a
	// repository name to start with a component as defined by DomainPortRegexp
	// and followed by an optional port.
	DomainComponentRegexp = Match(`(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`)

	// DomainRegexp defines the structure of potential domain components
	// that may be part of image names. This is purposely a subset of what is
	// allowed by DNS to ensure backwards compatibility with Docker image
	// names.
	DomainRegexp = Sequence(
		DomainComponentRegexp, Literal(`.`), DomainComponentRegexp,
		Optional(Repeated(Literal(`.`), DomainComponentRegexp)))

	// DomainPortRegexp defines the structure of potential domain components
	// that may be part of image names. This is purposely a subset of what is
	// allowed by DNS to ensure backwards compatibility with Docker image
	// names followed by an optional port part.
	DomainPortRegexp = Sequence(
		DomainRegexp,
		Optional(Literal(`:`), Match(`[0-9]+`)))

	// TagRegexp matches valid tag names. From docker/docker:graph/tags.go.
	TagRegexp = Match(`[\w][\w.-]{0,127}`)

	// AnchoredTagRegexp matches valid tag names, anchored at the start and
	// end of the matched string.
	AnchoredTagRegexp = Anchored(TagRegexp)

	// DigestRegexp matches valid digests.
	DigestRegexp = Match(`[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}`)

	// anchoredDigestRegexp matches valid digests, anchored at the start and
	// end of the matched string.
	anchoredDigestRegexp = Anchored(DigestRegexp)

	// RepositoryRegexp is the format of a repository ppart of references.
	RepositoryRegexp = Sequence(
		NameComponentRegexp,
		Optional(Repeated(RepositorySeparatorRegexp, NameComponentRegexp)))

	// AnchoredNameRegexp is used to parse a name value, capturing the
	// domain and trailing components.
	AnchoredNameRegexp = Anchored(
		Optional(Capture(DomainPortRegexp), RepositorySeparatorRegexp),
		Capture(RepositoryRegexp))

	// CapturedArtefactVersionRegexp is used to parse an artefact version sped
	// consisting of a repository part and an optional version part
	CapturedArtefactVersionRegexp = Sequence(
		Capture(RepositoryRegexp),
		CapturedVersionRegexp)

	// AnchoredArtefactVersionRegexp is used to parse artefact versions.
	AnchoredArtefactVersionRegexp = Anchored(CapturedArtefactVersionRegexp)

	// CapturedVersionRegexp described the version part of a reference
	CapturedVersionRegexp = Sequence(
		Optional(Literal(TagSeparator), Capture(TagRegexp)),
		Optional(Literal(DigestSeparator), Capture(DigestRegexp)))

	// ErrorCheckRegexp matches even wrong tags and/or digests
	ErrorCheckRegexp = Anchored(
		Optional(Capture(Match(".*?")), Literal("::")),
		Capture(Match(".*?")),
		Optional(Literal(TagSeparator), Capture(Match(".*?"))),
		Optional(Literal(DigestSeparator), Capture(Match(".*?"))))

	////////////////////////////////////////////////////////////////////////////
	// now the various full flegded artefact flavors

	// ReferenceRegexp is the full supported format of a reference. The regexp
	// is anchored and has capturing groups for name, tag, and digest
	// components.
	ReferenceRegexp = Anchored(
		Optional(Optional(CapturedSchemeRegexp), Capture(DomainPortRegexp), RepositorySeparatorRegexp),
		CapturedArtefactVersionRegexp)

	// DockerLibraryReferenceRegexp is a shortend docker library reference
	DockerLibraryReferenceRegexp = Anchored(
		Capture(NameComponentRegexp),
		CapturedVersionRegexp)

	// DockerReferenceRegexp is a shortend docker reference
	DockerReferenceRegexp = Anchored(
		Capture(NameComponentRegexp, RepositorySeparatorRegexp, NameComponentRegexp),
		CapturedVersionRegexp)

	TypedRepoRegexp = Anchored(
		Capture(TypeRegexp), Literal("::"),
		Optional(CapturedSchemeRegexp), Capture(DomainPortRegexp))

	TypedReferenceRegexp = Anchored(
		Capture(TypeRegexp), Literal("::"),
		Optional(CapturedSchemeRegexp), Capture(DomainPortRegexp),
		Optional(RepositorySeparatorRegexp, Optional(CapturedArtefactVersionRegexp)))

	TypedGenericReferenceRegexp = Anchored(
		Optional(Capture(TypeRegexp), Literal("::")),
		Capture(Match(".*?"), Match("[^:]")), Match(RepositorySeparator+RepositorySeparator),
		Optional(CapturedArtefactVersionRegexp))

	// Unused

	// IdentifierRegexp is the format for string identifier used as a
	// content addressable identifier using sha256. These identifiers
	// are like digests without the algorithm, since sha256 is used.
	IdentifierRegexp = Match(`([a-f0-9]{64})`)

	// ShortIdentifierRegexp is the format used to represent a prefix
	// of an identifier. A prefix may be used to match a sha256 identifier
	// within a list of trusted identifiers.
	ShortIdentifierRegexp = Match(`([a-f0-9]{6,64})`)

	// AnchoredIdentifierRegexp is used to check or match an
	// identifier value, anchored at start and end of string.
	AnchoredIdentifierRegexp = Anchored(IdentifierRegexp)
)
