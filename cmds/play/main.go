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

package main

import (
	"encoding/json"
	"fmt"

	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common"
	compdescv2 "github.com/gardener/ocm/pkg/ocm/compdesc/versions/v2"
	"github.com/gardener/ocm/pkg/runtime"
)

type Resources struct {
	*ResourceOptionsList `json:",inline"`
	*ResourceOptions     `json:",inline"`
}

// ResourceOptions contains options that are used to describe a resource
type ResourceOptions struct {
	compdescv2.Resource `json:",inline"`
	Input               *common.BlobInput `json:"input,omitempty"`
}

// ResourceOptionList contains a list of options that are used to describe a resource.
type ResourceOptionsList struct {
	Resources []json.RawMessage `json:"resources"`
}

type Abstract interface {
}

type Other struct {
	Other string `json:"other,omitempty"`
}

func main() {
	var a Abstract

	data := `
type: test
resources:
  - input:
      type: test
`

	data1 := `
type: test
`
	data2 := `
resources:
  - input:
      type: test
`
	a = &Resources{}

	err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(data2), a)
	fmt.Printf("err %s\n", err)

	var other map[string]interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal([]byte(data2), &other)
	fmt.Printf("err %s\n", err)

	_ = data
	_ = data1
	_ = data2
}
