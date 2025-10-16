package git

import (
	giturls "github.com/chainguard-dev/git-urls"
	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/git"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`

	// Repository is the Git Repository URL
	Repository string `json:"repository"`

	// Ref is the Git Ref to check out.
	// If empty, the default HEAD (remotes/origin/HEAD) of the remote is used.
	Ref string `json:"ref,omitempty"`

	// Commit is the Git Commit to check out.
	// If empty, the default HEAD of the Ref is used.
	Commit string `json:"commit,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(repository, ref, commit string) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		Repository: repository,
		Ref:        ref,
		Commit:     commit,
	}
}

func (s *Spec) Validate(fldPath *field.Path, _ inputs.Context, _ string) field.ErrorList {
	var allErrs field.ErrorList

	if path := fldPath.Child("repository"); s.Repository == "" {
		allErrs = append(allErrs, field.Invalid(path, s.Repository, "no repository"))
	} else {
		if _, err := giturls.Parse(s.Repository); err != nil {
			allErrs = append(allErrs, field.Invalid(path, s.Repository, err.Error()))
		}
	}

	if ref := fldPath.Child("ref"); s.Ref != "" {
		if err := plumbing.ReferenceName(s.Ref).Validate(); err != nil {
			allErrs = append(allErrs, field.Invalid(ref, s.Ref, "invalid ref"))
		}
	}

	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	blob, err := git.BlobAccess(
		git.WithURL(s.Repository),
		git.WithRef(s.Ref),
		git.WithCommit(s.Commit),
		git.WithCredentialContext(ctx),
		git.WithLoggingContext(ctx),
		git.WithCachingContext(ctx),
	)
	if err != nil {
		return nil, "", err
	}
	return blob, "", nil
}
