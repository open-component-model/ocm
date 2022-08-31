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

package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
)

func TransferComponents(printer common.Printer, closure TransportClosure, repo ocm.Repository, prefix string, all bool, tgt ocm.Repository, handler transferhandler.TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	if printer == nil {
		printer = common.NewPrinter(nil)
	}

	lister := repo.ComponentLister()
	if lister == nil {
		return errors.ErrNotSupported("ComponentLister")
	}
	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}
	comps, err := lister.GetComponents(prefix, all)
	if err != nil {
		return err
	}
	list := errors.ErrListf("component transport")
	for _, c := range comps {
		comp, err := repo.LookupComponent(c)
		if list.Addf(printer, err, "component %s", c) == nil {
			printer.Printf("transferring component %q...\n", c)
			subp := printer.AddGap("  ")
			vers, err := comp.ListVersions()
			if list.Addf(subp, err, "list versions for %s", c) == nil {
				for _, v := range vers {
					meta := &compdesc.ElementMeta{Name: c, Version: v}
					cv, h, err := handler.TransferVersion(repo, nil, meta)
					if err != nil {
						list.Addf(subp, err, "version %s", v)
						continue
					}
					if cv != nil {
						list.Addf(subp, TransferVersion(subp, closure, cv, tgt, h), "")
					}
				}
			}
		}
	}
	return list.Result()
}
