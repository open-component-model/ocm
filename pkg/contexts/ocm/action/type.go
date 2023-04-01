// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/action/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ActionSpec = cpi.ActionSpec

type ActionResult = cpi.ActionResult

type ActionType cpi.ActionType

////////////////////////////////////////////////////////////////////////////////

func EncodeActionSpec(s ActionSpec) ([]byte, error) {
	return cpi.EncodeActionSpec(s, runtime.DefaultJSONEncoding)
}

func DecodeActionSpec(data []byte) (ActionSpec, error) {
	return cpi.DecodeActionSpec(data, runtime.DefaultYAMLEncoding)
}

func EncodeActionResult(s ActionResult) ([]byte, error) {
	return cpi.EncodeActionResult(s, runtime.DefaultJSONEncoding)
}

func DecodeActionResult(data []byte) (ActionResult, error) {
	return cpi.DecodeActionResult(data, runtime.DefaultYAMLEncoding)
}

func SupportedActionVersions(name string) []string {
	return cpi.SupportedActionVersions(name)
}
