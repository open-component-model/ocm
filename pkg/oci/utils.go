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

package oci

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gardener/ocm/pkg/errors"
)

type OCIRef struct {
	Host       string
	Port       int
	Repository string
	Reference  string
}

func (r OCIRef) HostPort() string {
	if r.Port > 0 {
		return fmt.Sprintf("%s:%d", r.Host, r.Port)
	}
	return r.Host
}

func ParseOCIReference(ref string) (OCIRef, error) {
	idx := strings.Index(ref, "/")
	oci := OCIRef{}
	if idx < 0 {
		return OCIRef{}, errors.ErrInvalid("oci reference", ref)
	}
	host := ref[:idx]
	rest := ref[idx+1:]
	idx = strings.Index(host, ":")
	if idx < 0 {
		oci.Host = host
	} else {
		oci.Host = host[:idx]
		port, err := strconv.ParseInt(ref[idx+1:], 10, 32)
		if err != nil {
			return OCIRef{}, err
		}
		oci.Port = int(port)
	}
	idx = strings.Index(rest, ":")
	if idx < 0 {
		return OCIRef{}, errors.ErrInvalid("oci reference", ref)
	}
	oci.Repository = rest[:idx]
	oci.Reference = rest[idx+1:]
	return oci, nil
}
