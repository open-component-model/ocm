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

package ocm

type TransferOptions interface {
}

type DefaultTransferOptions struct {
	recursive        bool
	resourcesByValue bool
	sourcesByValue   bool
}

func (o *DefaultTransferOptions) SetRecursive(recursive bool) {
	o.recursive = recursive
}

func (o *DefaultTransferOptions) SetResourcesByValue(resourcesByValue bool) {
	o.resourcesByValue = resourcesByValue
}

func (o *DefaultTransferOptions) SetSourcesByValue(sourcesByValue bool) {
	o.sourcesByValue = sourcesByValue
}

func (o *DefaultTransferOptions) IsRecursive() bool {
	return o.recursive
}

func (o *DefaultTransferOptions) IsResourcesByValue() bool {
	return o.resourcesByValue
}

func (o *DefaultTransferOptions) IsSourcesByValue() bool {
	return o.sourcesByValue
}

type TransferOption interface {
	Apply(TransferOptions)
}

///////////////////////////////////////////////////////////////////////////////

type RecursiveTransferOption interface {
	SetRecursive(bool)
	IsRecursive() bool
}

type recursiveOption struct {
	recursive bool
}

func (o *recursiveOption) Apply(to TransferOptions) {
	to.(RecursiveTransferOption).SetRecursive(o.recursive)
}

func RecursiveTransfer(recursive bool) TransferOption {
	return &recursiveOption{
		recursive: recursive,
	}
}

///////////////////////////////////////////////////////////////////////////////

type ResourcesByValueTransferOption interface {
	SetResourcesByValue(bool)
	IsResourcesByValue() bool
}

type resourcesByValueOption struct {
	flag bool
}

func (o *resourcesByValueOption) Apply(to TransferOptions) {
	to.(ResourcesByValueTransferOption).SetResourcesByValue(o.flag)
}

func ResourcesByValueTransfer(flag bool) TransferOption {
	return &resourcesByValueOption{
		flag: flag,
	}
}

///////////////////////////////////////////////////////////////////////////////

type SourcesByValueTransferOption interface {
	SetSourcesByValue(bool)
	IsSourcesByValue() bool
}

type sourcesByValueOption struct {
	flag bool
}

func (o *sourcesByValueOption) Apply(to TransferOptions) {
	to.(SourcesByValueTransferOption).SetSourcesByValue(o.flag)
}

func SourcesByValueTransfer(flag bool) TransferOption {
	return &sourcesByValueOption{
		flag: flag,
	}
}
