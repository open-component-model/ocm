// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
)

// TransferComponentVersionWithHandler uses the specified transfer handler to control
// the transfer process.
func TransferComponentVersionWithHandler(pr common.Printer, cv ocm.ComponentVersionAccess, tgt ocm.Repository, handler transferhandler.TransferHandler) error {
	return TransferVersion(pr, nil, cv, tgt, handler)
}

// TransferComponentVersion uses the transfer handler based on the given options to control
// the transfer process. The default handler is the standard handler.
func TransferComponentVersion(printer common.Printer, cv ocm.ComponentVersionAccess, tgt ocm.Repository, optlist ...transferhandler.TransferOption) error {
	h, err := NewTransferHandler(optlist...)
	if err != nil {
		return err
	}
	return TransferComponentVersionWithHandler(printer, cv, tgt, h)
}
