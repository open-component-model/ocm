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

package docker

import (
	"github.com/containerd/containerd/remotes/docker"
)

var (
	ContextWithRepositoryScope           = docker.ContextWithRepositoryScope
	ContextWithAppendPullRepositoryScope = docker.ContextWithAppendPullRepositoryScope
	NewInMemoryTracker                   = docker.NewInMemoryTracker
	NewDockerAuthorizer                  = docker.NewDockerAuthorizer
	WithAuthClient                       = docker.WithAuthClient
	WithAuthHeader                       = docker.WithAuthHeader
	WithAuthCreds                        = docker.WithAuthCreds
)

type (
	Errors            = docker.Errors
	StatusTracker     = docker.StatusTracker
	Status            = docker.Status
	StatusTrackLocker = docker.StatusTrackLocker
)

func ConvertHosts(hosts docker.RegistryHosts) RegistryHosts {
	return func(host string) ([]RegistryHost, error) {
		list, err := hosts(host)
		if err != nil {
			return nil, err
		}
		result := make([]RegistryHost, len(list))
		for i, v := range list {
			result[i] = RegistryHost{
				Client:       v.Client,
				Authorizer:   v.Authorizer,
				Host:         v.Host,
				Scheme:       v.Scheme,
				Path:         v.Path,
				Capabilities: HostCapabilities(v.Capabilities),
				Header:       v.Header,
			}
		}
		return result, nil
	}
}
