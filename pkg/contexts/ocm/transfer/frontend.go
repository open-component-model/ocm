// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

// TransferComponentVersion uses the specified transfer handler to control
// the transfer process.
func TransferComponentVersion(pr common.Printer, cv ocm.ComponentVersionAccess, tgt ocm.Repository, handler transferhandler.TransferHandler) error {
	return TransferVersion(pr, nil, cv, tgt, handler)
}

// StandardTransferComponentVersion uses the standard transfer handler and its options to control
// the transfer process.
func StandardTransferComponentVersion(printer common.Printer, cv ocm.ComponentVersionAccess, tgt ocm.Repository, optlist ...transferhandler.TransferOption) error {
	opts, err := standard.New(optlist...)
	if err != nil {
		return err
	}
	return TransferVersion(printer, nil, cv, tgt, opts)
}
