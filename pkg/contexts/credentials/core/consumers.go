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

package core

import (
	"sync"
)

type _consumers struct {
	sync.RWMutex
	data map[string]*_consumer
}

func newConsumers() *_consumers {
	return &_consumers{
		data: map[string]*_consumer{},
	}
}

func (c *_consumers) Get(id ConsumerIdentity) *_consumer {
	c.RLock()
	defer c.RUnlock()
	return c.data[string(id.Key())]
}

func (c *_consumers) Set(id ConsumerIdentity, creds CredentialsSource) {
	c.Lock()
	defer c.Unlock()
	c.data[string(id.Key())] = &_consumer{
		identity:    id,
		credentials: creds,
	}
}

func (c *_consumers) Match(pattern ConsumerIdentity, m IdentityMatcher) *_consumer {
	c.RLock()
	defer c.RUnlock()
	var found *_consumer
	var cur ConsumerIdentity
	for _, s := range c.data {
		if m(pattern, cur, s.identity) {
			found = s
			cur = s.identity
		}
	}
	return found
}

type _consumer struct {
	identity    ConsumerIdentity
	credentials CredentialsSource
}

func (c *_consumer) GetCredentials() CredentialsSource {
	return c.credentials
}
