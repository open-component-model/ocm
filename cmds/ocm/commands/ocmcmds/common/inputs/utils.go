// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package inputs

import (
	"fmt"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/errors"
)

func FileInfo(ctx clictx.Context, path string, inputFilePath string) (os.FileInfo, string, error) {
	var err error
	var inputInfo os.FileInfo

	fs := ctx.FileSystem()
	inputPath, err := GetPath(ctx, path, inputFilePath)
	if err != nil {
		return nil, "", err
	}
	inputInfo, err = fs.Stat(inputPath)
	if err != nil {
		return nil, "", errors.Wrapf(err, "input path %q", inputPath)
	}
	return inputInfo, inputPath, nil
}

func GetBaseDir(fs vfs.FileSystem, filePath string) (string, error) {
	var wd string
	if len(filePath) == 0 {
		// default to working directory if no input filePath is given
		var err error
		wd, err = fs.Getwd()
		if err != nil {
			return "", fmt.Errorf("unable to read current working directory: %w", err)
		}
	} else {
		wd = filepath.Dir(filePath)
	}
	return wd, nil
}

func GetPath(ctx clictx.Context, path string, inputFilePath string) (string, error) {
	fs := ctx.FileSystem()
	if path == "" {
		return "", fmt.Errorf("path attribute required")
	}
	if filepath.IsAbs(path) {
		return path, nil
	} else {
		wd, err := GetBaseDir(fs, inputFilePath)
		if err != nil {
			return "", err
		}

		return filepath.Join(wd, path), nil
	}
}
