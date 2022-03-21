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

package vfsattr

import (
	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const ATTR_KEY = "github.com/mandelsoft/vfs"

var _osfs = osfs.New()

func Get(ctx datacontext.Context) vfs.FileSystem {
	v := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if v == nil {
		return _osfs
	}
	fs, _ := v.(vfs.FileSystem)
	return fs

}

func Set(ctx datacontext.Context, fs vfs.FileSystem) {
	ctx.GetAttributes().SetAttribute(ATTR_KEY, fs)
}
