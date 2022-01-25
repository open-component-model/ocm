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
	"strings"
)

const VersionSeparator = "/"

type VersionedTypedObject interface {
	TypedObject
	GetKind() string
	GetVersion() string
}

func TypeName(args ...string) string {
	if len(args) == 1 {
		return args[0]
	}
	if len(args) == 2 {
		if args[1] == "" {
			return args[0]
		}
		return args[0] + VersionSeparator + args[1]
	}
	panic("invlid call to TypeName, one or two arguments required")
}

type ObjectVersionedType ObjectType

// NewVersionedObjectType creates an ObjectVersionedType value
func NewVersionedObjectType(args ...string) ObjectVersionedType {
	return ObjectVersionedType{Type: TypeName(args...)}
}

// GetType returns the type of the object.
func (t ObjectVersionedType) GetType() string {
	return t.Type
}

// SetType sets the type of the object.
func (t *ObjectVersionedType) SetType(typ string) {
	t.Type = typ
}

// GetKind returns the kind of the object.
func (v ObjectVersionedType) GetKind() string {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		return t
	}
	return t[:i]
}

// SetKind sets the kind of the object.
func (v *ObjectVersionedType) SetKind(kind string) {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		v.Type = kind
	} else {
		v.Type = kind + t[i:]
	}
}

// GetVersion returns the version of the object.
func (v ObjectVersionedType) GetVersion() string {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		return "v1"
	}
	return t[i+1:]
}

// SetVersion sets the version of the object.
func (v *ObjectVersionedType) SetVersion(version string) {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		if version != "" {
			v.Type = v.Type + VersionSeparator + version
		}
	} else {
		if version != "" {
			v.Type = t[:i] + VersionSeparator + version
		} else {
			v.Type = t[:i]
		}
	}
}
