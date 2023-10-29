// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package elements contains builders
// for elements of a component version,
// aka resources, sources and references.
// The package itself contains builders
// for resource/source metadata and component
// references.
//
// The sub package artifactblob contains
// packages for building technology specific
// blob resources/sources. When added to a component
// version they result in local blobs.
//
// The sub package artifactaccess contains
// packages for building technology specific
// access specification based referential
// resources/sources. When added to a component
// version they result in external access
// specifications.
package elements
