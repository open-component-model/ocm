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

package spiff

import (
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

type Options struct {
	standard.Options
	script []byte
	fs     vfs.FileSystem
}

var _ ScriptOption = (*Options)(nil)
var _ ScriptFilesystemOption = (*Options)(nil)

func (o *Options) SetScript(data []byte) {
	o.script = data
}

func (o *Options) GetScript() []byte {
	return o.script
}

func (o *Options) SetScriptFilesystem(fs vfs.FileSystem) {
	o.fs = fs
}

func (o *Options) GetScriptFilesystem() vfs.FileSystem {
	return o.fs
}

///////////////////////////////////////////////////////////////////////////////

type ScriptOption interface {
	SetScript(data []byte)
	GetScript() []byte
}

type scriptOption struct {
	source string
	script func() ([]byte, error)
}

func (o *scriptOption) Apply(to transferhandler.TransferOptions) error {
	script, err := o.script()
	if err != nil {
		return err
	}
	_, err = spiffing.New().Unmarshal(o.source, script)
	if err != nil {
		return err
	}
	to.(ScriptOption).SetScript(script)
	return nil
}

func Script(data []byte) transferhandler.TransferOption {
	return &scriptOption{
		source: "script",
		script: func() ([]byte, error) { return data, nil },
	}
}

func ScriptByFile(path string, fss ...vfs.VFS) transferhandler.TransferOption {
	fs := osfs.New()
	if len(fss) > 0 {
		fs = fss[0]
	}
	return &scriptOption{
		source: path,
		script: func() ([]byte, error) { return vfs.ReadFile(fs, path) },
	}
}

///////////////////////////////////////////////////////////////////////////////

type ScriptFilesystemOption interface {
	SetScriptFilesystem(fs vfs.FileSystem)
	GetScriptFilesystem() vfs.FileSystem
}

type filesystemOption struct {
	fs vfs.FileSystem
}

func (o *filesystemOption) Apply(to transferhandler.TransferOptions) error {
	to.(ScriptFilesystemOption).SetScriptFilesystem(o.fs)
	return nil
}

func ScriptFilesystem(fs vfs.FileSystem) transferhandler.TransferOption {
	return &filesystemOption{
		fs: fs,
	}
}
