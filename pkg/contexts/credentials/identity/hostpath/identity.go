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

package hostpath

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
)

// IDENTITY_TYPE is the identity of this matcher.
const IDENTITY_TYPE = "hostpath"

// ID_HOSTNAME is a hostname.
const ID_HOSTNAME = "hostname"

// ID_PORT is a port.
const ID_PORT = "port"

// ID_PATHPREFIX is the path prefix below the host.
const ID_PATHPREFIX = "pathprefix"

func init() {
	cpi.RegisterStandardIdentityMatcher(IDENTITY_TYPE, IdentityMatcher(""), `Host and path based credential matcher

This matcher works on the following properties:

- *<code>`+ID_HOSTNAME+`</code>* (required): the hostname of a server
- *<code>`+ID_PORT+`</code>* (optional): the port of a server
- *<code>`+ID_PATHPREFIX+`</code>* (optional): a path prefix to match. The 
  element with the most matching path components is selected (separator is <code>/</code>).
`)
}

func IdentityMatcher(identityType string) cpi.IdentityMatcher {
	return func(pattern, cur, id core.ConsumerIdentity) bool {
		if identityType != "" && pattern[identityType] != "" && id[identityType] != "" && pattern[identityType] != id[identityType] {
			return false
		}
		if pattern[ID_HOSTNAME] != "" && pattern[ID_HOSTNAME] != id[ID_HOSTNAME] {
			return false
		}

		if pattern[ID_PORT] != "" {
			if id[ID_PORT] != "" && id[ID_PORT] != pattern[ID_PORT] {
				return false
			}
		}

		if pattern[ID_PATHPREFIX] != "" {
			if id[ID_PATHPREFIX] != "" {
				if len(id[ID_PATHPREFIX]) > len(pattern[ID_PATHPREFIX]) {
					return false
				}
				pcomps := strings.Split(pattern[ID_PATHPREFIX], "/")
				icomps := strings.Split(id[ID_PATHPREFIX], "/")
				if len(icomps) > len(pcomps) {
					return false
				}
				for i := range icomps {
					if pcomps[i] != icomps[i] {
						return false
					}
				}
			}
		} else {
			if id[ID_PATHPREFIX] != "" {
				return false
			}
		}

		// ok now it basically matches, check against current match
		if len(cur) == 0 {
			return true
		}

		if cur[ID_HOSTNAME] == "" && id[ID_HOSTNAME] != "" {
			return true
		}
		if cur[ID_PORT] == "" && (id[ID_PORT] != "" && pattern[ID_PORT] != "") {
			return true
		}

		if len(cur[ID_PATHPREFIX]) < len(id[ID_PATHPREFIX]) {
			return true
		}
		return false
	}
}
