// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package processing

type operation interface {
	process(e interface{}) (operationResult, bool)
}

type operationResult = []interface{}

type explode ExplodeFunction

func (this explode) process(e interface{}) (operationResult, bool) {
	return this(e), true
}

type mapper MappingFunction

func (this mapper) process(e interface{}) (operationResult, bool) {
	return operationResult{this(e)}, true
}

type filter FilterFunction

func (this filter) process(e interface{}) (operationResult, bool) {
	if this(e) {
		return operationResult{e}, true
	}
	return nil, false
}
