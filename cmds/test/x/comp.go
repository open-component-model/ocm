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

package x

import (
	"fmt"
	"os"

	compdesc2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

func CheckErr(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %s\n ", fmt.Sprintf(msg, args), err)
		os.Exit(1)
	}
}

func C() (err error) {
	defer compdesc2.CatchConversionError(&err)
	C1()
	return
}

func C1() {
	compdesc2.ThrowConversionError(fmt.Errorf("occured"))
}

func compTest() {
	data, err := os.ReadFile("component-descriptor.yaml")
	CheckErr(err, "read")
	cd, err := compdesc2.Decode(data)
	CheckErr(err, "decode")

	raw, err := cd.RepositoryContexts[0].GetRaw()
	CheckErr(err, "raw ctx")
	fmt.Printf("ctx: %s\n", string(raw))
	_ = cd
	data, err = compdesc2.Encode(cd)
	CheckErr(err, "marshal")
	fmt.Printf("%s\n", string(data))

	err = C()
	fmt.Printf("catched error %s\n", err)
}
