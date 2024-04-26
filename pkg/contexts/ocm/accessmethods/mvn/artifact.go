package mvn

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/mimeutils"
)

// Artifact holds the typical Maven coordinates groupId, artifactId, version and packaging.
// https://maven.apache.org/ref/3.9.6/maven-core/artifact-handlers.html
type Artifact struct {
	GroupId    string
	ArtifactId string
	Version    string
	Classifier string
	Extension  string
	//	Type       string
	//	Packaging  string
}

// GAV returns the GAV coordinates of the Maven Artifact.
func (a *Artifact) GAV() string {
	return a.GroupId + ":" + a.ArtifactId + ":" + a.Version
}

// String returns the GAV coordinates of the Maven Artifact.
func (a *Artifact) String() string {
	return a.GAV()
}

// GavPath returns the Maven repository path.
func (a *Artifact) GavPath() string {
	return a.GroupPath() + "/" + a.ArtifactId + "/" + a.Version
}

// FileName returns the Maven Artifact's GAV-name with classifier and extension.
// Which is equal to the URL-path of the artifact in the repository.
// Default extension is jar.
func (a *Artifact) FileName() string {
	path := a.GavPath() + "/" + a.FilePrefix()
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
	return baseUrl + "/" + a.FileName()
}

// GroupPath returns GroupId with `/` instead of `.`.
func (a *Artifact) GroupPath() string {
	return strings.ReplaceAll(a.GroupId, ".", "/")
}

func (a *Artifact) FilePrefix() string {
	return a.ArtifactId + "-" + a.Version
}

// Purl returns the Package URL of the Maven Artifact.
func (a *Artifact) Purl() string {
	return "pkg:maven/" + a.GroupId + "/" + a.ArtifactId + "@" + a.Version
}

// ClassifierExtensionFrom extracts the classifier and extension from the filename (without any path prefix).
func (a *Artifact) ClassifierExtensionFrom(filename string) *Artifact {
	// TODO should work with pos (path.Basename)?!?
	s := strings.TrimPrefix(filename, a.FilePrefix())
	if strings.HasPrefix(s, "-") {
		s = strings.TrimPrefix(s, "-")
		i := strings.Index(s, ".")
		if i < 0 {
			panic("no extension after classifier found in filename: " + filename)
		}
		a.Classifier = s[:i]
		s = strings.TrimPrefix(s, a.Classifier)
	} else {
		a.Classifier = ""
	}
	a.Extension = strings.TrimPrefix(s, ".")
	return a
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

// ArtifactFromHint creates new Artifact from accessspec-hint. See 'GetReferenceHint'.
func ArtifactFromHint(gav string) *Artifact {
	parts := strings.Split(gav, ":")
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
