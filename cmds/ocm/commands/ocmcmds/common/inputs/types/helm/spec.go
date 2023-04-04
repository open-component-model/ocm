// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm/loader"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
)

type Spec struct {
	// PathSpec hold the path that points to the helm chart file
	cpi.PathSpec `json:",inline"`
	Version      string `json:"version,omitempty"`
	Repository   string `json:"repository,omitempty"`
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

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (accessio.TemporaryBlobAccess, string, error) {
	_, inputPath, err := inputs.FileInfo(ctx, s.Path, info.InputFilePath)
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
		vers = info.ComponentVersion.GetVersion()
	}
	blob, err := helm.SynthesizeArtifactBlob(inputPath, ctx.FileSystem())
	if err != nil {
		return nil, "", err
	}

	return blob, ociartifact.Hint(info.ComponentVersion, chart.Name(), s.Repository, vers), err
}
