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
