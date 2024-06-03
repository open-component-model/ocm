package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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
	return TransferWithHandler(local.printer, cv, tgt, h)
}
