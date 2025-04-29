package transfer

import (
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/utils/misc"
)

type (
	// TransferOption if the interface for options given to transfer functions.
	// The can influence the behaviour of the transfer process by configuring
	// appropriate transfer handlers.
	TransferOption = transferhandler.TransferOption

	// TransferOptions is the target interface for consumers of
	// a TransferOption.
	TransferOptions = transferhandler.TransferOptions

	// TransferHandler controls the transfer of component versions.
	// It can be used to control the value transport of sources and resources
	// on artifact level (by providing specific handling for dedicated artifact attributes),
	// the concrete re/source transfer step, and the way how
	// nested component version are transported.
	// There are two implementations delivered as part of the OCM library:
	//   - package transferhandler.standard: able to select recursive transfer
	//     general value artifact transport.
	//   - package transferhandler.spiff: controls transfer using a spiff script.
	// Custom implementations can be used to gain fine-grained control
	// over the transfer process, whose general flow is handled by
	// a uniform Transfer function.
	TransferHandler = transferhandler.TransferHandler
)

// Local options do not relate to the transfer handler, but directly to the
// processing logic. They are formal transferhandler options to be passable to
// the option list but apply themselves only for the localOptions object.
// To distinguish them from transferhandler options, they do NOT implement
// the transferhandler.TransferOptionsCreator interface.
type localOptions struct {
	printer misc.Printer
}

func (opts *localOptions) Eval(optlist ...transferhandler.TransferOption) error {
	for _, o := range optlist {
		if _, ok := o.(transferhandler.TransferOptionsCreator); !ok {
			err := o.ApplyTransferOption(opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// WithPrinter provides a explicit printer object. By default,
// a non-printing printer will be used.
func WithPrinter(p misc.Printer) transferhandler.TransferOption {
	return &localOptions{
		printer: p,
	}
}

func (l *localOptions) ApplyTransferOption(options TransferOptions) error {
	if t, ok := options.(*localOptions); ok {
		if l.printer != nil {
			t.printer = l.printer
		}
	}
	return nil
}
