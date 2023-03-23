// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package standard

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Options struct {
	recursive        bool
	resourcesByValue bool
	sourcesByValue   bool
	keepGlobalAccess bool
	overwrite        bool
	resolver         ocm.ComponentVersionResolver
}

var (
	_ ResourcesByValueOption = (*Options)(nil)
	_ SourcesByValueOption   = (*Options)(nil)
	_ RecursiveOption        = (*Options)(nil)
	_ ResolverOption         = (*Options)(nil)
	_ KeepGlobalAccessOption = (*Options)(nil)
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

func (o *Options) SetKeepGlobalAccess(keepGlobalAccess bool) {
	o.keepGlobalAccess = keepGlobalAccess
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

func (o *Options) IsKeepGlobalAccess() bool {
	return o.keepGlobalAccess
}

func (o *Options) GetResolver() ocm.ComponentVersionResolver {
	return o.resolver
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
		overwrite: utils.GetOptionFlag(args...),
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
		recursive: utils.GetOptionFlag(args...),
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
		flag: utils.GetOptionFlag(args...),
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
		flag: utils.GetOptionFlag(args...),
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

///////////////////////////////////////////////////////////////////////////////

type KeepGlobalAccessOption interface {
	SetKeepGlobalAccess(bool)
	IsKeepGlobalAccess() bool
}

type keepGlobalOption struct {
	flag bool
}

func (o *keepGlobalOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(KeepGlobalAccessOption); ok {
		eff.SetKeepGlobalAccess(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported("resolver")
	}
}

func KeepGlobalAccess(args ...bool) transferhandler.TransferOption {
	return &keepGlobalOption{
		flag: utils.GetOptionFlag(args...),
	}
}
