package transfer

import (
	"context"

	"ocm.software/ocm/api/ocm"
	common "ocm.software/ocm/api/utils/misc"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// TransferWithHandler uses the specified transfer handler to control
// the transfer process.
func TransferWithHandler(pr common.Printer, cv ocm.ComponentVersionAccess, tgt ocm.Repository, handler TransferHandler) error {
	return TransferVersion(pr, nil, cv, tgt, handler)
}

// Transfer uses the transfer handler based on the given options to control
// the transfer process. The default handler is the standard handler.
func Transfer(cv ocm.ComponentVersionAccess, tgt ocm.Repository, optlist ...TransferOption) error {
	h, err := NewTransferHandler(optlist...)
	if err != nil {
		return err
	}
	var local localOptions
	err = local.Eval(optlist...)
	if err != nil {
		return err
	}
	return TransferVersionWithContext(common.WithPrinter(context.Background(), local.printer), nil, cv, tgt, h)
}

// TransferWithContext uses the transfer handler based on the given options to control
// the transfer process. The default handler is the standard handler.
func TransferWithContext(ctx context.Context, cv ocm.ComponentVersionAccess, tgt ocm.Repository, optlist ...TransferOption) error {
	h, err := NewTransferHandler(optlist...)
	if err != nil {
		return err
	}
	var local localOptions
	err = local.Eval(optlist...)
	if err != nil {
		return err
	}
	if local.printer != nil {
		ctx = common.WithPrinter(ctx, local.printer)
	}
	return TransferVersionWithContext(ctx, nil, cv, tgt, h)
}
