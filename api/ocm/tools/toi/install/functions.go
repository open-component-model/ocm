package install

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/spiff/yaml"
	"ocm.software/ocm/api/ocm"
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
		if len(arguments) < 1 || len(arguments) > 2 {
			return info.Error("credential name and optional property name argument required")
		}
		name, ok := arguments[0].(string)
		if !ok {
			return info.Error("credential name argument must be string")
		}
		creds := credvals[name]
		if creds == nil {
			return info.Error("credential %q not found", name)
		}
		if len(arguments) == 1 {
			val := map[string]spiffing.Node{}
			for n, v := range creds {
				val[n] = yaml.NewNode(v, "credential")
			}
			return val, info, true
		}
		key, ok := arguments[1].(string)
		if !ok {
			return info.Error("credential property argument must be string")
		}
		if key == "*" || key == "" {
			if len(creds) > 1 {
				return info.Error("there are multiple credential properties")
			}
			for _, v := range creds {
				return v, info, true
			}
			return "", info, true
		}
		v, ok := creds[key]
		if !ok {
			return info.Error("there is no credential property %q", key)
		}
		return v, info, true
	}
}

func spiffHasCredentials(credvals CredentialValues) dynaml.Function {
	return func(arguments []interface{}, binding dynaml.Binding) (interface{}, dynaml.EvaluationInfo, bool) {
		var info dynaml.EvaluationInfo
		if len(arguments) < 1 || len(arguments) > 2 {
			return info.Error("credential name and optional property name argument required")
		}
		name, ok := arguments[0].(string)
		if !ok {
			return info.Error("credential name argument must be string")
		}
		creds := credvals[name]
		if creds == nil || len(arguments) == 1 {
			return creds != nil, info, true
		}

		key, ok := arguments[1].(string)
		if !ok {
			return info.Error("credential property argument must be string")
		}

		if key == "*" || key == "" {
			return len(creds) == 1, info, true
		}
		_, ok = creds[key]
		return ok, info, true
	}
}
