// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func Load(name string, fs vfs.FileSystem) (*chart.Chart, error) {
	fi, err := fs.Stat(name)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return LoadDir(fs, name)
	}
	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return loader.LoadArchive(file)
}
