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

package standard

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
)

type Options struct {
	recursive        bool
	resourcesByValue bool
	sourcesByValue   bool
	overwrite        bool
}

var _ ResourcesByValueOption = (*Options)(nil)
var _ SourcesByValueOption = (*Options)(nil)
var _ RecursiveOption = (*Options)(nil)

func (o *Options) SetOverwrite(overwrite bool) {
	o.overwrite = overwrite
}

func (o *Options) SetRecursive(recursive bool) {
	o.recursive = recursive
}

func (o *Options) SetResourcesByValue(resourcesByValue bool) {
	o.resourcesByValue = resourcesByValue
}

func (o *Options) SetSourcesByValue(sourcesByValue bool) {
	o.sourcesByValue = sourcesByValue
}

func (o *Options) IsOverwrite() bool {
	return o.overwrite
}

func (o *Options) IsRecursive() bool {
	return o.recursive
}

func (o *Options) IsResourcesByValue() bool {
	return o.resourcesByValue
}

func (o *Options) IsSourcesByValue() bool {
	return o.sourcesByValue
}

///////////////////////////////////////////////////////////////////////////////

func GetFlag(args ...bool) bool {
	flag := len(args) == 0
	for _, f := range args {
		if f {
			flag = true
			break
		}
	}
	return flag
}

///////////////////////////////////////////////////////////////////////////////

type OverwriteOption interface {
	SetOverwrite(bool)
	IsOverwrite() bool
}

type overwriteOption struct {
	overwrite bool
}

func (o *overwriteOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	to.(OverwriteOption).SetOverwrite(o.overwrite)
	return nil
}

func Overwrite(args ...bool) transferhandler.TransferOption {
	return &overwriteOption{
		overwrite: GetFlag(args...),
	}
}

///////////////////////////////////////////////////////////////////////////////

type RecursiveOption interface {
	SetRecursive(bool)
	IsRecursive() bool
}

type recursiveOption struct {
	recursive bool
}

func (o *recursiveOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	to.(RecursiveOption).SetRecursive(o.recursive)
	return nil
}

func Recursive(args ...bool) transferhandler.TransferOption {
	return &recursiveOption{
		recursive: GetFlag(args...),
	}
}

///////////////////////////////////////////////////////////////////////////////

type ResourcesByValueOption interface {
	SetResourcesByValue(bool)
	IsResourcesByValue() bool
}

type resourcesByValueOption struct {
	flag bool
}

func (o *resourcesByValueOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	to.(ResourcesByValueOption).SetResourcesByValue(o.flag)
	return nil
}

func ResourcesByValue(args ...bool) transferhandler.TransferOption {
	return &resourcesByValueOption{
		flag: GetFlag(args...),
	}
}

///////////////////////////////////////////////////////////////////////////////

type SourcesByValueOption interface {
	SetSourcesByValue(bool)
	IsSourcesByValue() bool
}

type sourcesByValueOption struct {
	flag bool
}

func (o *sourcesByValueOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	to.(SourcesByValueOption).SetSourcesByValue(o.flag)
	return nil
}

func SourcesByValue(args ...bool) transferhandler.TransferOption {
	return &sourcesByValueOption{
		flag: GetFlag(args...),
	}
}
