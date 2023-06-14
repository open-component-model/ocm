// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package ppi provides the plugin programming interface.
// The ppi can be used by plugin developers as support. It reduces the command line interface to a corresponding Go
// interface. In other words, if the developer implements this Go plugin interface, the ppi automatically provides a
// corresponding command line interface.
package ppi
