package maven

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/utils/blobaccess"
	mavenblob "ocm.software/ocm/api/utils/blobaccess/maven"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

type Spec struct {
	cpi.PathSpec `json:",inline"`
	// RepoUrl defines the url from which the artifact is downloaded.
	RepoUrl string `json:"repoUrl,omitempty"`

	maven.Coordinates `json:",inline"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(repoUrl, groupId, artifactId, version string, classifier, extension *string) *Spec {
	return &Spec{
		PathSpec: cpi.NewPathSpec(TYPE, ""),
		RepoUrl:  repoUrl,
		Coordinates: *maven.NewCoordinates(groupId, artifactId, version,
			maven.WithOptionalClassifier(classifier),
			maven.WithOptionalExtension(extension)),
	}
}

func NewForFilePath(filePath, groupId, artifactId, version string, classifier, extension *string) *Spec {
	return &Spec{
		PathSpec: cpi.NewPathSpec(TYPE, filePath),
		RepoUrl:  "",
		Coordinates: *maven.NewCoordinates(groupId, artifactId, version,
			maven.WithOptionalClassifier(classifier),
			maven.WithOptionalExtension(extension)),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	var allErrs field.ErrorList
	if s.RepoUrl == "" {
		allErrs = s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	} else {
		if s.Path != "" {
			pathField := fldPath.Child("path")
			allErrs = append(allErrs, field.Forbidden(pathField, "only path or repoUrl can be specified, not both"))
		}
	}
	if s.ArtifactId == "" {
		pathField := fldPath.Child("artifactId")
		allErrs = append(allErrs, field.Invalid(pathField, s.ArtifactId, "no artifact id"))
	}
	if s.GroupId == "" {
		pathField := fldPath.Child("groupId")
		allErrs = append(allErrs, field.Invalid(pathField, s.GroupId, "no group id"))
	}
	if s.Version == "" {
		pathField := fldPath.Child("version")
		allErrs = append(allErrs, field.Invalid(pathField, s.GroupId, "no version"))
	}

	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	var repo *maven.Repository
	var err error

	fs := ctx.FileSystem()
	if s.Path != "" {
		inputInfo, inputPath, err := inputs.FileInfo(ctx, s.Path, info.InputFilePath)
		if err != nil {
			return nil, "", err
		}
		if !inputInfo.IsDir() {
			return nil, "", fmt.Errorf("maven file repository must be a directory")
		}
		repo = maven.NewFileRepository(inputPath, fs)
	} else {
		repo, err = maven.NewUrlRepository(s.RepoUrl, fs)
		if err != nil {
			return nil, "", err
		}
	}
	access, err := mavenblob.BlobAccessForCoords(repo, &s.Coordinates,
		mavenblob.WithCredentialContext(ctx),
		mavenblob.WithLoggingContext(ctx),
		mavenblob.WithCachingFileSystem(vfsattr.Get(ctx)),
	)

	if s.IsPackage() {
		return access, s.GAV(), err
	}

	return access, "", err
}
