// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/clisupport"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm/loader"
)

func ConfigHandler() clisupport.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig,
		options.PathOption, options.VersionOption)
}

type Spec struct {
	// PathSpec hold the path that points to the helm chart file
	cpi.PathSpec `json:",inline"`
	Version      string `json:"version,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path string) *Spec {
	return &Spec{
		PathSpec: cpi.NewPathSpec(TYPE, path),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	allErrs := s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	if s.Path != "" {
		path := fldPath.Child("path")
		inputInfo, filePath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(path, filePath, err.Error()))
		} else if !inputInfo.Mode().IsDir() && !inputInfo.Mode().IsRegular() {
			allErrs = append(allErrs, field.Invalid(path, filePath, "no regular file or directory"))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	_, inputPath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
	if err != nil {
		return nil, "", err
	}
	chart, err := loader.Load(inputPath, ctx.FileSystem())
	if err != nil {
		return nil, "", err
	}
	vers := chart.Metadata.Version
	if s.Version != "" {
		vers = s.Version
	}
	if vers == "" {
		vers = nv.GetVersion()
	}
	blob, err := helm.SynthesizeArtefactBlob(inputPath, ctx.FileSystem())
	if err != nil {
		return nil, "", err
	}
	name := chart.Name()
	hint := fmt.Sprintf("%s/%s:%s", nv.GetName(), name, vers)
	if name == "" {
		hint = fmt.Sprintf("%s:%s", nv.GetName(), vers)
	}
	return blob, hint, err
}

func AddConfig(opts clisupport.ConfigOptions, config clisupport.Config) error {
	if err := cpi.AddPathSpecConfig(opts, config); err != nil {
		return err
	}
	if v, ok := opts.GetValue(options.VersionOption.Name()); ok {
		config["version"] = v
	}
	return nil
}
