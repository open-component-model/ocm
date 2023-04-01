// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci_repository_prepare

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/action/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const Type = "oci.repository.prepare"

func init() {
	cpi.RegisterAction(Type, &ActionSpec{}, &ActionResult{}, "Prepare the usage of a repository in an OCI registry.")

	cpi.RegisterType(Type, "v1", cpi.NewActionTypeByProtoTypes(&ActionSpecV1{}, nil, &ActionResultV1{}, nil))
}

////////////////////////////////////////////////////////////////////////////////
// internal version

type ActionSpec = ActionSpecV1

type ActionResult = ActionResultV1

////////////////////////////////////////////////////////////////////////////////
// serialization formats

type ActionSpecV1 struct {
	runtime.VersionedTypedObject
	Hostname string `json:"hostname"`
}

func (s *ActionSpecV1) Selector() cpi.Selector {
	return cpi.Selector(s.Hostname)
}

type ActionResultV1 struct {
	cpi.CommonResult `json:",inline"`
}
