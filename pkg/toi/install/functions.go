// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/spiff/yaml"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

func NewFunctions(ctx ocm.Context, credvals CredentialValues) spiffing.Functions {
	funcs := spiffing.NewFunctions()

	funcs.RegisterFunction("getCredentials", spiffGetCredentials(credvals))
	funcs.RegisterFunction("hasCredentials", spiffHasCredentials(credvals))
	return funcs
}

func spiffGetCredentials(credvals CredentialValues) spiffing.Function {
	return func(arguments []interface{}, binding dynaml.Binding) (interface{}, dynaml.EvaluationInfo, bool) {
		var info dynaml.EvaluationInfo
		if len(arguments) != 1 {
			return info.Error("credential name argument required")
		}
		name, ok := arguments[0].(string)
		if !ok {
			return info.Error("credential name argument must be string")
		}
		creds := credvals[name]
		if creds == nil {
			return info.Error("credential %q not found", name)
		}
		val := map[string]spiffing.Node{}
		for n, v := range creds {
			val[n] = yaml.NewNode(v, "credential")
		}
		return val, info, true
	}
}

func spiffHasCredentials(credvals CredentialValues) dynaml.Function {
	return func(arguments []interface{}, binding dynaml.Binding) (interface{}, dynaml.EvaluationInfo, bool) {
		var info dynaml.EvaluationInfo
		if len(arguments) != 1 {
			return info.Error("credential name argument required")
		}
		name, ok := arguments[0].(string)
		if !ok {
			return info.Error("credential name argument must be string")
		}
		creds := credvals[name]
		return creds != nil, info, true
	}
}
