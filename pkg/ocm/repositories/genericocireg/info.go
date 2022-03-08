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

package genericocireg

import (
	"encoding/json"
	"fmt"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/ociutils"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg/componentmapping"
)

func init() {
	ociutils.RegisterInfoHandler(componentmapping.ComponentDescriptorConfigMimeType, &handler{})
}

type handler struct{}

func (h handler) Info(m cpi.ManifestAccess, config []byte) string {
	s := "component version:\n"
	acc := NewStateAccess(m)
	data, err := accessio.BlobData(acc.Get())
	if err != nil {
		return s + "  cannot read component descriptor: " + err.Error()
	}
	s += fmt.Sprintf("  descriptor:\n")
	var raw interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		s += "    " + string(data)
		return s + "  cannot get unmarshal component descriptor: " + err.Error()
	}

	form, err := json.MarshalIndent(raw, "  ", "    ")
	if err != nil {
		s += "    " + string(data)
		return s + "  cannot get marshal component descriptor: " + err.Error()
	}
	return s + string(form)
}
