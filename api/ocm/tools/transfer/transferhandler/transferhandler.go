package transferhandler

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
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
// to regular transfer handler options. This is used to infer an applicable
// transferhandler for the given option set.
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
	// UpdateVersion decides whether an update of volatile (non-signature relevant) parts of a CV should be updated.
	UpdateVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error)
	// EnforceTransport decides whether a component version should be transport as it is.
	// This controls whether transport is carried out
	// as if the component version were not present at the destination.
	EnforceTransport(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error)
	// OverwriteVersion decides whether a modification of non-volatile (signature relevant) parts of a CV should be updated.
	OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error)

	// TransferVersion decides on continuing with a component version (reference).
	TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.Reference, tgt ocm.Repository) (ocm.ComponentVersionAccess, TransferHandler, error)
	// TransferResource decides on the value transport of a resource.
	TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error)
	// TransferSource decides on the value transport of a source.
	TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error)

	// HandleTransferResource technically transfers a resource.
	HandleTransferResource(r ocm.ResourceAccess, m cpi.AccessMethod, hint string, t ocm.ComponentVersionAccess) error
	// HandleTransferSource technically transfers a source.
	HandleTransferSource(r ocm.SourceAccess, m cpi.AccessMethod, hint string, t ocm.ComponentVersionAccess) error
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
