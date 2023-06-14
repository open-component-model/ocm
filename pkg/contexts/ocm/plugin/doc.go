// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package plugin maps a Go plugin interface to a technical command line interface. Thereby, it allows to embed plugins
// and their provided functionality into the library. It is the basis for adapter implementations for accessmethods,
// downloaders, uploaders and blob handlers.
// Those adapters can be found in the subpackage "plugin" of their dedicated functional area packages (e.g.
// pkg/contexts/ocm/accessmethods/plugin).
package plugin
