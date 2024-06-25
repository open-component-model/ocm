package utils

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func CheckForUnknownFields(fldPath *field.Path, orig, accepted map[string]interface{}) field.ErrorList {
	allErrs := field.ErrorList{}

	for k, o := range orig {
		child := fldPath.Child(k)
		if a, ok := accepted[k]; ok {
			allErrs = append(allErrs, CheckForUnknown(child, o, a)...)
		} else {
			// this is a hack, empty lists or maps are not covered with omitempty in json annotation
			switch v := o.(type) {
			case map[string]interface{}:
				if len(v) == 0 {
					continue
				}
			case []interface{}:
				if len(v) == 0 {
					continue
				}
			}
			allErrs = append(allErrs, field.Forbidden(child, "unknown field"))
		}
	}
	return allErrs
}

func CheckForUnknownElements(fldPath *field.Path, orig, accepted []interface{}) field.ErrorList {
	allErrs := field.ErrorList{}
	for i, o := range orig {
		if i >= len(accepted) {
			allErrs = append(allErrs, field.Forbidden(fldPath, "unexpected list entry"))
		} else {
			allErrs = append(allErrs, CheckForUnknown(fldPath.Index(i), o, accepted[i])...)
		}
	}
	return allErrs
}

func CheckForUnknown(fldPath *field.Path, orig, accepted interface{}) field.ErrorList {
	allErrs := field.ErrorList{}
	switch a := accepted.(type) {
	case map[string]interface{}:
		if o, ok := orig.(map[string]interface{}); ok {
			allErrs = append(allErrs, CheckForUnknownFields(fldPath, o, a)...)
		} else {
			allErrs = append(allErrs, field.Forbidden(fldPath, "map expected"))
		}
	case []interface{}:
		if o, ok := orig.([]interface{}); ok {
			allErrs = append(allErrs, CheckForUnknownElements(fldPath, o, a)...)
		} else {
			allErrs = append(allErrs, field.Forbidden(fldPath, "list expected"))
		}
	default:
	}
	return allErrs
}
