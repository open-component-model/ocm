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
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type Unstructured struct {
	NotToMarshal string
	Object       json.RawMessage `json:"-"`
}

// MarshalJSON returns m as the JSON encoding of m.
func (m Unstructured) MarshalJSON() ([]byte, error) {
	return m.Object.MarshalJSON()
}

// UnmarshalJSON sets *m to a copy of data.
func (m *Unstructured) UnmarshalJSON(data []byte) error {
	return m.Object.UnmarshalJSON(data)
}

func ToUnstructured(payload interface{}) *Unstructured {
	data, err := json.Marshal(payload)
	if err != nil {
		panic("cannot marshal")
	}
	unstr := &Unstructured{}
	err = json.Unmarshal(data, unstr)
	if err != nil {
		panic("cannot unmarshal")
	}
	return unstr
}

////////////////////////////////////////////////////////////////////////////////

type Generic struct {
	Unstructured `json:",inline"`
}

func ToGeneric(payload interface{}) *Generic {
	return &Generic{*ToUnstructured(payload)}
}

type Spec struct {
	Repository Unstructured `json:"repository"`
	Value      string       `json:"value"`
}

////////////////////////////////////////////////////////////////////////////////

type SomeData struct {
	Type string `json:"type"`
}

////////////////////////////////////////////////////////////////////////////////

type Direct struct {
	Repositories Spec `json:"repositories"`
}

type Map struct {
	Repositories map[string]Spec `json:"repositories"`
}

////////////////////////////////////////////////////////////////////////////////

func CheckData(msg string, data []byte, expect string) {
	fmt.Printf("%s:\n", msg)
	fmt.Printf("  found:    %s\n", string(data))
	if string(data) != expect {
		fmt.Fprintf(os.Stderr, "  expected: %s\n", expect)
		os.Exit(1)
	}
}

func CheckValue(v reflect.Value) {
	if !v.IsValid() {
		panic("value not valid")
	}
}

func ReflectTest() {
	value := map[string]Spec{
		"entry": Spec{
			Repository: Unstructured{},
			Value:      "value",
		},
	}

	mv := reflect.ValueOf(value)
	ev := mv.MapIndex(reflect.ValueOf("entry"))
	CheckValue(ev)

	if !ev.CanAddr() {
		fmt.Printf("oops: map entry\n")
	}

	fv := ev.Field(0)
	CheckValue(fv)
	fmt.Printf("field type %s: %s\n", ev.Type().Field(0).Name, fv.Type().String())

	if !fv.CanAddr() {
		fmt.Printf("oops: field %s\n", ev.Type().Field(0).Name)
	}
}

func JsonTest() {
	ReflectTest()
	payload := &SomeData{Type: "someType"}

	u := ToUnstructured(payload)

	udata, err := json.Marshal(u)
	CheckErr(err, "marshal unstructured")
	CheckData("payload", udata, "{\"type\":\"someType\"}")

	spec := &Spec{
		Repository: *ToUnstructured(payload),
		Value:      "someValue",
	}

	sdata, err := json.Marshal(spec)
	CheckErr(err, "marshal spec")
	CheckData("spec", sdata, "{\"repository\":{\"type\":\"someType\"},\"value\":\"someValue\"}")

	direct := &Direct{
		Repositories: *spec,
	}

	ddata, err := json.Marshal(direct)
	CheckErr(err, "marshal direct")
	CheckData("direct", ddata, "{\"repositories\":{\"repository\":{\"type\":\"someType\"},\"value\":\"someValue\"}}")

	smap := &Map{
		Repositories: map[string]Spec{
			"entry": *spec,
		},
	}

	mdata, err := json.Marshal(smap)
	CheckErr(err, "marshal map")
	CheckData("map", mdata, "{\"repositories\":{\"entry\":{\"repository\":{\"type\":\"someType\"},\"value\":\"someValue\"}}}")

	nmap := &Map{}

	err = json.Unmarshal(mdata, nmap)
	CheckErr(err, "unmarshal map")
	if !reflect.DeepEqual(nmap, smap) {
		fmt.Printf("unmarshaled marhaled differs from origin\n")
	}

}
