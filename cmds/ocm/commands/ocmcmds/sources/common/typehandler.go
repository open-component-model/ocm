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

package common

import (
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
)

func Elem(e interface{}) *compdesc.Source {
	return e.(*common.Object).Element.(*compdesc.Source)
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	*common.TypeHandler
}

func NewTypeHandler(repo ocm.Repository, session ocm.Session, access ocm.ComponentVersionAccess, recursive bool) utils.TypeHandler {
	return common.NewTypeHandler(repo, session, access, recursive, func(access ocm.ComponentVersionAccess) compdesc.ElementAccessor {
		return access.GetDescriptor().Sources
	})
}
