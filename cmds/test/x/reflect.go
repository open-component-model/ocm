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
	"reflect"
)

type Getter interface {
	GetType() string
}
type IA interface {
	Getter
	IA()
}

type ia struct{}

func (ia) IA()             {}
func (ia) GetType() string { return "" }

func NewDefaultScheme(v interface{}) error {
	if v == nil {
		return fmt.Errorf("prototype must be given by pointer to interace (is nil)")
	}
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		return fmt.Errorf("prototype %T: must be given by pointer to interace (is not pointer)", v)
	}
	rt = rt.Elem()
	if rt.Kind() != reflect.Interface {
		return fmt.Errorf("prototype %T: must be given by pointer to interace (does not point to interface)", v)
	}
	fmt.Printf("var type: %s (%s)\n", rt, rt.Kind())
	return nil
}

func DoReflect() {
	var v IA
	var s ia

	rt := reflect.TypeOf(&v).Elem()
	fmt.Printf("var type: %s (%s)\n", rt, rt.Kind())

	fmt.Printf("check: %s\n", NewDefaultScheme(&v))
	fmt.Printf("check: %s\n", NewDefaultScheme(v))
	fmt.Printf("check: %s\n", NewDefaultScheme(s))
	fmt.Printf("check: %s\n", NewDefaultScheme(&s))

	v = ia{}
	fmt.Printf("check: %s\n", NewDefaultScheme(v))
	fmt.Printf("check: %s\n", NewDefaultScheme(&v))
	rv := reflect.ValueOf(v)
	fmt.Printf("var type: %s (%s)\n", rv.Type(), rv.Type().Kind())

	fmt.Printf("implements: %t\n", rv.Type().Implements(rt))
}
