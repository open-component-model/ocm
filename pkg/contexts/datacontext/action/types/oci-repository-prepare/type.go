// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci_repository_prepare

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/api"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const Type = "oci.repository.prepare"

func init() {
	api.RegisterAction(Type, &ActionSpec{}, &ActionResult{}, "Prepare the usage of a repository in an OCI registry.")

	api.RegisterType(Type, "v1", api.NewActionTypeByProtoTypes(&ActionSpecV1{}, nil, &ActionResultV1{}, nil))
}

////////////////////////////////////////////////////////////////////////////////
// internal version

type ActionSpec = ActionSpecV1

type ActionResult = ActionResultV1

func Spec(host string, repo string) *ActionSpec {
	return &ActionSpec{
		ObjectVersionedType: runtime.ObjectVersionedType{runtime.TypeName(Type, "v1")},
		Hostname:            host,
		Repository:          repo,
	}
}

func Result(msg string) *ActionResult {
	return &ActionResult{
		CommonResult: api.CommonResult{
			ObjectVersionedType: runtime.ObjectVersionedType{runtime.TypeName(Type, "v1")},
			Message:             msg,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// serialization formats

type ActionSpecV1 struct {
	runtime.ObjectVersionedType
	Hostname   string `json:"hostname"`
	Repository string `json:"repository"`
}

func (s *ActionSpecV1) Selector() api.Selector {
	return api.Selector(s.Hostname)
}

type ActionResultV1 struct {
	api.CommonResult `json:",inline"`
}
