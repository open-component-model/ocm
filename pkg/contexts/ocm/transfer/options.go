// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
)

type (
	TransferOption  = transferhandler.TransferOption
	TransferOptions = transferhandler.TransferOptions
)

// Local options do not relate to the transfer handler, but directly to the
// processing logic. They are formal transferhandler options to be passable to
// the option list but apply themselves only for the localOptions object.
// To distinguish them from transferhandler options, they do NOT implement
// the transferhandler.TransferOptionsCreator interface.
type localOptions struct {
	printer common.Printer
}

func (opts *localOptions) Eval(optlist ...transferhandler.TransferOption) error {
	var local localOptions
	for _, o := range optlist {
		if _, ok := o.(transferhandler.TransferOptionsCreator); !ok {
			err := o.ApplyTransferOption(&local)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// WithPrinter provides a explicit printer object. By default,
// a non-printing printer will be used.
func WithPrinter(p common.Printer) transferhandler.TransferOption {
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
