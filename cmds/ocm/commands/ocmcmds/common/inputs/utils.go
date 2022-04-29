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

package inputs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ForbidFileInfo(fldPath *field.Path, input *BlobInput) field.ErrorList {
	allErrs := ForbidFilePattern(fldPath, input)
	path := fldPath.Child("compress")
	if input.CompressWithGzip != nil {
		allErrs = append(allErrs, field.Required(path, fmt.Sprintf("compress option not possible for type %s", input.Type)))
	}
	return allErrs
}

func ForbidFilePattern(fldPath *field.Path, input *BlobInput) field.ErrorList {
	allErrs := field.ErrorList{}
	path := fldPath.Child("includeFiles")
	if input.IncludeFiles != nil {
		allErrs = append(allErrs, field.Required(path, fmt.Sprintf("includeFiles option not possble for type %s", input.Type)))
	}
	path = fldPath.Child("excludeFiles")
	if input.ExcludeFiles != nil {
		allErrs = append(allErrs, field.Required(path, fmt.Sprintf("excludeFiles option not possble for type %s", input.Type)))
	}
	path = fldPath.Child("preserveDir")
	if input.PreserveDir != nil {
		allErrs = append(allErrs, field.Required(path, fmt.Sprintf("preserveDir option not possble for type %s", input.Type)))
	}
	path = fldPath.Child("followSymlinks")
	if input.FollowSymlinks != nil {
		allErrs = append(allErrs, field.Required(path, fmt.Sprintf("followSymlinks option not possble for type %s", input.Type)))
	}

	return allErrs
}