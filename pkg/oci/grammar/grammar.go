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

var (
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
	// repository name to start with a component as defined by DomainRegexp
	// and followed by an optional port.
	DomainComponentRegexp = Match(`(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`)

	// DomainRegexp defines the structure of potential domain components
	// that may be part of image names. This is purposely a subset of what is
	// allowed by DNS to ensure backwards compatibility with Docker image
	// names.
	DomainRegexp = Sequence(
		DomainComponentRegexp,
		Optional(Repeated(Literal(`.`), DomainComponentRegexp)),
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
		Optional(Repeated(Literal(`/`), NameComponentRegexp)))

	// NameRegexp is the format for the name component of references. The
	// regexp has capturing groups for the domain and name part omitting
	// the separating forward slash from either.
	NameRegexp = Sequence(
		Optional(DomainRegexp, Literal(`/`)), RepositoryRegexp)

	// AnchoredNameRegexp is used to parse a name value, capturing the
	// domain and trailing components.
	AnchoredNameRegexp = Anchored(
		Optional(Capture(DomainRegexp), Literal(`/`)),
		Capture(NameComponentRegexp,
			Optional(Repeated(Literal(`/`), NameComponentRegexp))))

	// ArtefactVersionRegexp is used to parse an artefact version sped
	// consisting of a repository part and an optional version part
	ArtefactVersionRegexp = Sequence(
		Capture(RepositoryRegexp),
		Optional(Literal(":"), Capture(TagRegexp)),
		Optional(Literal("@"), Capture(DigestRegexp)))

	// AnchoredArtefactVersionRegexp is used to parse artefact versions.
	AnchoredArtefactVersionRegexp = Anchored(ArtefactVersionRegexp)

	// ReferenceRegexp is the full supported format of a reference. The regexp
	// is anchored and has capturing groups for name, tag, and digest
	// components.
	ReferenceRegexp = Anchored(Capture(NameRegexp),
		Optional(Literal(":"), Capture(TagRegexp)),
		Optional(Literal("@"), Capture(DigestRegexp)))

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
