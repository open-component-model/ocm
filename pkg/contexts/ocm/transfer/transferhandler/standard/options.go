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
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Options struct {
	recursive        bool
	resourcesByValue bool
	sourcesByValue   bool
	overwrite        bool
	resolver         ocm.ComponentVersionResolver
}

var (
	_ ResourcesByValueOption = (*Options)(nil)
	_ SourcesByValueOption   = (*Options)(nil)
	_ RecursiveOption        = (*Options)(nil)
	_ ResolverOption         = (*Options)(nil)
)

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

func (o *Options) SetResolver(resolver ocm.ComponentVersionResolver) {
	o.resolver = resolver
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

func (o *Options) GetResolver() ocm.ComponentVersionResolver {
	return o.resolver
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
	if eff, ok := to.(OverwriteOption); ok {
		eff.SetOverwrite(o.overwrite)
		return nil
	} else {
		return errors.ErrNotSupported("overwrite")
	}
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
	if eff, ok := to.(RecursiveOption); ok {
		eff.SetRecursive(o.recursive)
		return nil
	} else {
		return errors.ErrNotSupported("recursive")
	}
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
	if eff, ok := to.(ResourcesByValueOption); ok {
		eff.SetResourcesByValue(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported("resources by-value")
	}
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
	if eff, ok := to.(SourcesByValueOption); ok {
		eff.SetSourcesByValue(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported("sources by-value")
	}
}

func SourcesByValue(args ...bool) transferhandler.TransferOption {
	return &sourcesByValueOption{
		flag: GetFlag(args...),
	}
}

///////////////////////////////////////////////////////////////////////////////

type ResolverOption interface {
	GetResolver() ocm.ComponentVersionResolver
	SetResolver(ocm.ComponentVersionResolver)
}

type resolverOption struct {
	resolver ocm.ComponentVersionResolver
}

func (o *resolverOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(ResolverOption); ok {
		eff.SetResolver(o.resolver)
		return nil
	} else {
		return errors.ErrNotSupported("resolver")
	}
}

func Resolver(resolver ocm.ComponentVersionResolver) transferhandler.TransferOption {
	return &resolverOption{
		resolver: resolver,
	}
}
