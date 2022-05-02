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

package ocm

import (
	"fmt"
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/errors"
)

////////////////////////////////////////////////////////////////////////////////

func AssureTargetRepository(session Session, ctx Context, targetref string, format accessio.FileFormat, fss ...vfs.FileSystem) (Repository, error) {
	target, ref, err := session.DetermineRepository(ctx, targetref)
	if err != nil {
		if !errors.IsErrUnknown(err) || ref.Info == "" {
			return nil, err
		}
		if ref.Type == "" {
			ref.Type = format.String()
		}
		if ref.Type == "" {
			return nil, fmt.Errorf("ctf format type required to create ctf")
		}
		target, err = ctf.Create(ctx, accessobj.ACC_CREATE, ref.Info, 0770, accessio.PathFileSystem(accessio.FileSystem(fss...)))
		if err != nil {
			return nil, err
		}
		session.Closer(target)
	}
	return target, nil
}

type AccessMethodSource interface {
	AccessMethod() (AccessMethod, error)
}

// ResourceReader gets a Reader for a given resource/source access.
// It provides a Reader handling the Close contract for the access method
// by connecting the access method's Close method to the Readers Close method .
func ResourceReader(s AccessMethodSource) (io.ReadCloser, error) {
	meth, err := s.AccessMethod()
	if err != nil {
		return nil, err
	}
	r, err := meth.Reader()
	if err != nil {
		meth.Close()
		return nil, err
	}
	return accessio.AddCloser(r, meth, "access method"), nil
}

// ResourceData extracts the data for a given resource/source access.
// It handles the Close contract for the access method for a singular use.
func ResourceData(s AccessMethodSource) ([]byte, error) {
	meth, err := s.AccessMethod()
	if err != nil {
		return nil, err
	}
	defer meth.Close()
	return meth.Get()
}
