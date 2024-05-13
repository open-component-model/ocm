package mvn

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/mimeutils"
)

// Artifact holds the typical Maven coordinates groupId, artifactId, version. Optional also classifier and extension.
// https://maven.apache.org/ref/3.9.6/maven-core/artifact-handlers.html
type Artifact struct {
	// ArtifactId is the name of Maven (mvn) artifact.
	GroupId string `json:"groupId"`
	// ArtifactId is the name of Maven (mvn) artifact.
	ArtifactId string `json:"artifactId"`
	// Version of the Maven (mvn) artifact.
	Version string `json:"version"`
	// Classifier of the Maven (mvn) artifact.
	Classifier string `json:"classifier"`
	// Extension of the Maven (mvn) artifact.
	Extension string `json:"extension"`
}

// GAV returns the GAV coordinates of the Maven Artifact.
func (a *Artifact) GAV() string {
	return a.GroupId + ":" + a.ArtifactId + ":" + a.Version
}

// Serialize returns the Artifact as a string (GroupId:ArtifactId:Version:Classifier:Extension).
func (a *Artifact) Serialize() string {
	return a.GroupId + ":" + a.ArtifactId + ":" + a.Version + ":" + a.Classifier + ":" + a.Extension
}

// String returns the GAV coordinates of the Maven Artifact.
func (a *Artifact) String() string {
	return a.GAV()
}

// GavPath returns the Maven repository path.
func (a *Artifact) GavPath() string {
	return a.GroupPath() + "/" + a.ArtifactId + "/" + a.Version
}

// FilePath returns the Maven Artifact's GAV-name with classifier and extension.
// Which is equal to the URL-path of the artifact in the repository.
// Default extension is jar.
func (a *Artifact) FilePath() string {
	path := a.GavPath() + "/" + a.FileNamePrefix()
	if a.Classifier != "" {
		path += "-" + a.Classifier
	}
	if a.Extension != "" {
		path += "." + a.Extension
	} else {
		path += ".jar"
	}
	return path
}

func (a *Artifact) Url(baseUrl string) string {
	return baseUrl + "/" + a.FilePath()
}

// GroupPath returns GroupId with `/` instead of `.`.
func (a *Artifact) GroupPath() string {
	return strings.ReplaceAll(a.GroupId, ".", "/")
}

func (a *Artifact) FileNamePrefix() string {
	return a.ArtifactId + "-" + a.Version
}

// Purl returns the Package URL of the Maven Artifact.
func (a *Artifact) Purl() string {
	return "pkg:maven/" + a.GroupId + "/" + a.ArtifactId + "@" + a.Version
}

// ClassifierExtensionFrom extracts the classifier and extension from the filename (without any path prefix).
func (a *Artifact) ClassifierExtensionFrom(filename string) (*Artifact, error) {
	// TODO should work with both (path.Basename)?!?
	s := strings.TrimPrefix(filename, a.FileNamePrefix())
	if strings.HasPrefix(s, "-") {
		s = strings.TrimPrefix(s, "-")
		i := strings.Index(s, ".")
		if i < 0 {
			return nil, fmt.Errorf("no extension after classifier found in filename: %s", filename)
		}
		a.Classifier = s[:i]
		s = strings.TrimPrefix(s, a.Classifier)
	} else {
		a.Classifier = ""
	}
	a.Extension = strings.TrimPrefix(s, ".")
	return a, nil
}

// MimeType returns the MIME type of the Maven Artifact based on the file extension.
// Default is application/x-tgz.
func (a *Artifact) MimeType() string {
	m := mimeutils.TypeByExtension("." + a.Extension)
	if m != "" {
		return m
	}
	return mime.MIME_TGZ
}

// Copy creates a new Artifact with the same values.
func (a *Artifact) Copy() *Artifact {
	return &Artifact{
		GroupId:    a.GroupId,
		ArtifactId: a.ArtifactId,
		Version:    a.Version,
		Classifier: a.Classifier,
		Extension:  a.Extension,
	}
}

// DeSerialize creates an Artifact from it's serialized form (see Artifact.Serialize).
func DeSerialize(serializedArtifact string) *Artifact {
	parts := strings.Split(serializedArtifact, ":")
	if len(parts) < 3 {
		return nil
	}
	artifact := &Artifact{
		GroupId:    parts[0],
		ArtifactId: parts[1],
		Version:    parts[2],
	}
	if len(parts) >= 4 {
		artifact.Classifier = parts[3]
	}
	if len(parts) >= 5 {
		artifact.Extension = parts[4]
	}
	return artifact
}

// IsResource returns true if the filename is not a checksum or signature file.
func IsResource(fileName string) bool {
	if strings.HasSuffix(fileName, ".asc") {
		return false
	}
	if strings.HasSuffix(fileName, ".md5") {
		return false
	}
	if strings.HasSuffix(fileName, ".sha1") {
		return false
	}
	if strings.HasSuffix(fileName, ".sha256") {
		return false
	}
	if strings.HasSuffix(fileName, ".sha512") {
		return false
	}
	return true
}
