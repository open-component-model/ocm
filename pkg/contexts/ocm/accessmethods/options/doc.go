// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package options defines standard options and option types usable
// to provide CLI options used to dynamically orchestrate arbitrary
// access specifications. These options have a predefined meaning and
// are shared among various access methods.
//
// The options and types are registered at a global registry.
// This registry is also used by the plugin adapter to
// map option requests from plugins to implementations.
package options
