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
