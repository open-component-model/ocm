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

	"github.com/open-component-model/ocm/pkg/common/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/toi/drivers/docker"
	"github.com/open-component-model/ocm/pkg/toi/install"
)

func MustParseYaml(data string) json.RawMessage {
	var m map[string]interface{}
	err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &m)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(data)
}

func exec() {

	cfg := MustParseYaml(`
special: config
`)
	templ := MustParseYaml(`
parameters:
   username: admin
   password: (( &merge ))
`)
	scheme := MustParseYaml(`
type: object
required:
  - parameters
additionalProperties: false
properties:
  parameters:
    type: object
    required:
    - password
    additionalProperties: false
    properties:
      username:
        type: string
      password:
        type: string
`)

	params := MustParseYaml(`
parameters:
   password: supersecret
`)

	spec := &install.PackageSpecification{
		Template: templ,
		Scheme:   scheme,
		Executors: []install.Executor{
			{
				Image:  &install.Image{Ref: "inst"},
				Config: cfg,
				Outputs: map[string]string{
					"test": "bla",
				},
			},
		},
	}

	_ = spec

	config.Configure("")

	octx := ocm.DefaultContext()

	sess := ocm.NewSession(nil)
	ref, err := sess.EvaluateRef(octx, "ghcr.io/mandelsoft/cnudie//github.com/mandelsoft/ocmdemoinstaller:0.0.1-dev")
	CheckErr(err, "inst component")
	r, err := install.Execute(&docker.Driver{}, "install", metav1.NewIdentity("demoinstaller"), params, octx, ref.Version, nil)
	//r, err := install.ExecuteAction(&docker.Driver{}, "install", spec, params, nil, nil)

	CheckErr(err, "execute")

	fmt.Printf("result: %s\n", r.Outputs["bla"])
}
