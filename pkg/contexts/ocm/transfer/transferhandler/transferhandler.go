// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transferhandler

import (
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

const KIND_TRANSFEROPTION = "transfer option"

// TransferHandlerOptions is the general type for a dedicated kind of option set
// holding options for a dedicated type of transfer handler.
// Different transfer handler implementations may use different
// concrete option sets.
// To support option sets with an overlapping set of accepted options,
// it is not possible to use fixed structs with explicit fields
// for every option as target for the apply methods of the option
// objects. Therefore, every option uses its own option type interfaces.
// The option's setters/getters MUST be interfaces, which may be implemented
// by different option set types to enable the option implementations
// to work with different options sets.
//
// For example the spiff transfer handler options include all the standard
// handler options.
type TransferHandlerOptions interface {
	TransferOptions
	NewTransferHandler() (TransferHandler, error)
}

// TransferOptions is the general interface used by TransferOption implementations
// as target to apply themselves to.
type TransferOptions interface{}

// TransferOptionsCreator is an optional interface for a TransferOption.
// The option may provide a default TransferOptions object if it applies
// to regular transfer handler options. THis is used to infer an applicable
// transfer hander for the gicven option set.
// Options not intended for the transfer handler MUST NOT implement this
// interface to not hamper the handler detection process.
type TransferOptionsCreator interface {
	NewOptions() TransferHandlerOptions
}

// TransferOption is an option used to configure the transfer process.
// This interface is used by transfer operations for the optional list
// of given options.
// Every option decides on its own, whether it is applicable to the given target,
// which is given by the generic TransferOptions interface.
// There are two kinds of options:
//   - handler specific options additionally implement the TransferOptionsCreator
//     interface.
//   - operation related option do not implement this interface.
type TransferOption interface {
	ApplyTransferOption(TransferOptions) error
}

type optionsPointer[P any] interface {
	TransferHandlerOptions
	*P
}

// SpecializedOptionsCreator is the base implementation for options objects
// for specialized transfer handlers.
type SpecializedOptionsCreator[P optionsPointer[T], T any] struct{}

func (o SpecializedOptionsCreator[P, T]) NewOptions() TransferHandlerOptions {
	var opts T
	return P(&opts)
}

////////////////////////////////////////////////////////////////////////////////

// TransferHandler controls the transfer of component versions.
// It can be used to control the value transport of sources and resources
// on artifact level (by providing specific handling for dedicated artifact attributes),
// the concrete re/source transfer step, and the way how
// nested component version are transported.
// There are two implementations delivered as part of the OCM library:
//   - package transferhandler.standard: able to select recursive transfer
//     general value artifact transport.
//   - package transferhandler.spiff: controls transfer using a spiff script.
type TransferHandler interface {
	OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error)

	TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.ComponentReference, tgt ocm.Repository) (ocm.ComponentVersionAccess, TransferHandler, error)
	TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error)
	TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error)

	// HandleTransferResource technically transfers a resource.
	// The access method must be closed by this method.
	HandleTransferResource(r ocm.ResourceAccess, m ocm.AccessMethod, hint string, t ocm.ComponentVersionAccess) error
	// HandleTransferSource technically transfers a source.
	// The access method must be closed by this method.
	HandleTransferSource(r ocm.SourceAccess, m ocm.AccessMethod, hint string, t ocm.ComponentVersionAccess) error
}

func ApplyOptions(set TransferOptions, opts ...TransferOption) error {
	list := errors.ErrListf("transfer options")
	for _, o := range opts {
		list.Add(o.ApplyTransferOption(set))
	}
	return list.Result()
}

func From(ctx config.ContextProvider, opts TransferOptions) error {
	_, err := ctx.ConfigContext().ApplyTo(-1, opts)
	return err
}

func BoolP(b bool) *bool {
	return &b
}

func AsBool(b *bool) bool {
	return b != nil && *b
}
