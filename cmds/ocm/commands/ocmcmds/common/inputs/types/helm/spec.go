package helm

import (
	"github.com/mandelsoft/goutils/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/helm"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
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
			switch {
			case err != nil:
				allErrs = append(allErrs, field.Invalid(path, filePath, err.Error()))
			case !inputInfo.Mode().IsRegular():
				allErrs = append(allErrs, field.Invalid(path, filePath, "caCertFile is no regular file"))
			default:
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

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blob blobaccess.BlobAccess, hint string, err error) {
	path := s.Path
	if s.HelmRepository == "" {
		_, inputPath, err := inputs.FileInfo(ctx, path, info.InputFilePath)
		if err != nil {
			return nil, "", errors.Wrapf(err, "cannot handle input path %q", s.Path)
		}
		path = inputPath
	}
	vers := s.Version
	override := true
	if vers == "" {
		vers = info.ComponentVersion.GetVersion()
		override = false
	}

	blob, name, vers, err := helm.BlobAccess(path,
		helm.WithContext(ctx),
		helm.WithFileSystem(ctx.FileSystem()),
		helm.WithPrinter(ctx.Printer()),
		helm.WithVersionOverride(vers, override),
		helm.WithCACert(s.CACert),
		helm.WithCACertFile(s.CACertFile),
		helm.WithHelmRepository(s.HelmRepository),
	)
	if err != nil {
		return nil, "", err
	}
	hint = ociartifact.Hint(info.ComponentVersion, name, s.Repository, vers)
	return blob, hint, err
}

func (s *Spec) GetInputVersion(ctx inputs.Context) string {
	return s.Version
}
