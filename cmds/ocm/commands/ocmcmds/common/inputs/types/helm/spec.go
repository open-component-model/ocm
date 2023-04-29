// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	ocihelm "github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/helm"
	"github.com/open-component-model/ocm/pkg/helm/credentials"
	"github.com/open-component-model/ocm/pkg/helm/loader"
)

type Spec struct {
	// PathSpec hold the path that points to the helm chart file
	cpi.PathSpec   `json:",inline"`
	HelmRepository string `json:"helmRepository,omitempty"`
	Version        string `json:"version,omitempty"`
	Repository     string `json:"repository,omitempty"`
	CACert         string `json:"caCert,omitempty"`
	CACertFile     string `json:"caCertFile,omitempty"`
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
		if s.HelmRepository == "" {
			path := fldPath.Child("path")
			inputInfo, filePath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(path, filePath, err.Error()))
			} else if !inputInfo.Mode().IsDir() && !inputInfo.Mode().IsRegular() {
				allErrs = append(allErrs, field.Invalid(path, filePath, "no regular file or directory"))
			}
		}
	}
	if s.HelmRepository != "" {
		path := fldPath.Child("version")
		if s.Version == "" {
			allErrs = append(allErrs, field.Required(path, "required if field 'helmRepository' is set"))
		}

		if s.CACertFile != "" {
			path = fldPath.Child("caCertFile")
			inputInfo, filePath, err := inputs.FileInfo(ctx, s.CACertFile, inputFilePath)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(path, filePath, err.Error()))
			} else if !inputInfo.Mode().IsRegular() {
				allErrs = append(allErrs, field.Invalid(path, filePath, "caCertFile is no regular file"))
			} else {
				_, err = LoadCertificateBundle(s.CACertFile, ctx.FileSystem())
				if err != nil {
					allErrs = append(allErrs, field.Invalid(path, s.CACertFile, err.Error()))
				}
			}
		}

		if s.CACert != "" {
			path = fldPath.Child("caCert")
			_, err := LoadCertificateBundleFromData([]byte(s.CACert))
			if err != nil {
				allErrs = append(allErrs, field.Invalid(path, s.CACertFile, err.Error()))
			}
		}
	} else {
		if s.CACertFile != "" {
			path := fldPath.Child("caCertFile")
			allErrs = append(allErrs, field.Forbidden(path, "only possible together with helmRepository"))
		}
		if s.CACert != "" {
			path := fldPath.Child("caCert")
			allErrs = append(allErrs, field.Forbidden(path, "only possible together with helmRepository"))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (accessio.TemporaryBlobAccess, string, error) {
	var chartLoader loader.Loader
	if s.HelmRepository == "" {
		_, inputPath, err := inputs.FileInfo(ctx, s.Path, info.InputFilePath)
		if err != nil {
			return nil, "", errors.Wrapf(err, "cannot handle input path %q", s.Path)
		}
		chartLoader = loader.VFSLoader(inputPath, ctx.FileSystem())

	} else {
		cert := []byte(s.CACert)
		if s.CACertFile != "" {
			_, certPath, err := inputs.FileInfo(ctx, s.CACertFile, info.InputFilePath)
			if err != nil {
				return nil, "", err
			}
			cert, err = vfs.ReadFile(ctx.FileSystem(), certPath)
			if err != nil {
				return nil, "", errors.Wrapf(err, "cannot read root certificates from %q", s.CACertFile)
			}
		}

		acc, err := helm.DownloadChart(ctx.StdOut(), s.Path, s.Version, s.HelmRepository,
			helm.WithCredentials(credentials.GetCredentials(ctx, s.HelmRepository)),
			helm.WithRootCert([]byte(cert)))
		if err != nil {
			return nil, "", errors.Wrapf(err, "cannot download chart %s:%s from %s", s.Path, s.Version, s.HelmRepository)
		}
		chartLoader = loader.AccessLoader(acc)
	}
	defer chartLoader.Close()

	chart, err := chartLoader.Chart()
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
	blob, err := ocihelm.SynthesizeArtifactBlob(chartLoader)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot synthesize artifact blob")
	}

	return blob, ociartifact.Hint(info.ComponentVersion, chart.Name(), s.Repository, vers), err
}

func (s *Spec) GetInputVersion(ctx inputs.Context) string {
	return s.Version
}
