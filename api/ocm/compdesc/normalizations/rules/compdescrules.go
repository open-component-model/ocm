package rules

import (
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/tech/signing"
)

func IgnoreLabelsWithoutSignature(v interface{}) bool {
	if m, ok := v.(map[string]interface{}); ok {
		if sig, ok := m["signing"]; ok {
			if sig != nil {
				return sig != "true" && sig != true
			}
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

var LabelExcludes = signing.ExcludeEmpty{signing.DynamicArrayExcludes{
	ValueChecker: IgnoreLabelsWithoutSignature,
	Continue: signing.MapIncludes{
		"name":    signing.NoExcludes{},
		"version": signing.NoExcludes{},
		"value":   signing.NoExcludes{},
		"signing": signing.NoExcludes{},
	},
}}

////////////////////////////////////////////////////////////////////////////////

func MapResourcesWithNoneAccess(v interface{}) interface{} {
	return MapResourcesWithAccessType(
		compdesc.IsNoneAccessKind,
		func(v interface{}) interface{} {
			m := v.(map[string]interface{})
			delete(m, "digest")
			return m
		},
		v,
	)
}

func IgnoreResourcesWithNoneAccess(v interface{}) bool {
	return CheckIgnoreResourcesWithAccessType(compdesc.IsNoneAccessKind, v)
}

func IgnoreResourcesWithAccessType(t string) func(v interface{}) bool {
	return func(v interface{}) bool {
		return CheckIgnoreResourcesWithAccessType(func(k string) bool { return k == t }, v)
	}
}

func CheckIgnoreResourcesWithAccessType(t func(string) bool, v interface{}) bool {
	access := v.(map[string]interface{})["access"]
	if access == nil {
		return true
	}
	typ := access.(map[string]interface{})["type"]
	if s, ok := typ.(string); ok {
		return t(s)
	}
	return false
}

func MapResourcesWithAccessType(t func(string) bool, m func(interface{}) interface{}, v interface{}) interface{} {
	access := v.(map[string]interface{})["access"]
	if access == nil {
		return v
	}
	typ := access.(map[string]interface{})["type"]
	if s, ok := typ.(string); ok {
		if t(s) {
			return m(v)
		}
	}
	return v
}
