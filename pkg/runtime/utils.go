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

package runtime

import (
	"reflect"
	"sort"
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
)

func MustProtoType(proto interface{}) reflect.Type {
	t, err := ProtoType(proto)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func ProtoType(proto interface{}) (reflect.Type, error) {
	if proto == nil {
		return nil, errors.New("prototype required")
	}
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("prototype %q must be a struct", t)
	}
	return t, nil
}

func TypedObjectFactory(proto TypedObject) func() TypedObject {
	return func() TypedObject { return reflect.New(MustProtoType(proto)).Interface().(TypedObject) }
}

func TypeNames(scheme Scheme) []string {
	types := []string{}
	for t := range scheme.KnownTypes() {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func KindNames(scheme Scheme) []string {
	types := []string{}
	for t := range scheme.KnownTypes() {
		if !strings.Contains(t, VersionSeparator) {
			types = append(types, t)
		}
	}
	sort.Strings(types)
	return types
}
