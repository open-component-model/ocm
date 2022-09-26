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

package identity

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
)

// CONSUMER_TYPE is the OCT registry type.
const CONSUMER_TYPE = "OCIRegistry"

// ID_TYPE is the type field of a consumer identity.
const ID_TYPE = cpi.ID_TYPE

// ID_HOSTNAME is the hostname of an OCT repository.
const ID_HOSTNAME = hostpath.ID_HOSTNAME

// ID_PORT is the port number of an OCT repository.
const ID_PORT = hostpath.ID_PORT

// ID_PATHPREFIX is the artefact prefix.
const ID_PATHPREFIX = hostpath.ID_PATHPREFIX

// ID_SCHEME is the scheme prefix.
const ID_SCHEME = hostpath.ID_SCHEME

func init() {
	cpi.RegisterStandardIdentityMatcher(CONSUMER_TYPE, IdentityMatcher, `OCI registry credential matcher

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`)
}

var identityMatcher = hostpath.IdentityMatcher(CONSUMER_TYPE)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}
