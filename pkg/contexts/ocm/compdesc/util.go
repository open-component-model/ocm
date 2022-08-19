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

package compdesc

type conversionError struct {
	error
}

func (e conversionError) Error() string {
	return "conversion error: " + e.error.Error()
}

func CatchConversionError(errp *error) {
	if r := recover(); r != nil {
		if je, ok := r.(conversionError); ok {
			*errp = je
		} else {
			panic(r)
		}
	}
}

func Validate(desc *ComponentDescriptor) error {
	data, err := Encode(desc)
	if err != nil {
		return err
	}
	_, err = Decode(data)
	return err
}
