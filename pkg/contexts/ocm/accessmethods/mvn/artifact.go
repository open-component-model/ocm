package mvn

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/mime"
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

// FileName returns the Maven Artifact's name with classifier and extension.
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

func (a *Artifact) ClassifierExtensionFrom(filename string) *Artifact {
	s := strings.TrimPrefix(filename, a.FilePrefix())
	if strings.HasPrefix(s, "-") {
		s = strings.TrimPrefix(s, "-")
		a.Classifier = s[:strings.Index(s, ".")]
		s = strings.TrimPrefix(s, a.Classifier)
	}
	a.Extension = strings.TrimPrefix(s, ".")
	return a
}

// MimeType returns the MIME type of the Maven Artifact based on the file extension.
// Default is application/x-tgz.
func (a *Artifact) MimeType() string {
	switch a.Extension {
	case "jar":
		return mime.MIME_JAR
	case "json", "module":
		return mime.MIME_JSON
	case "pom", "xml":
		return mime.MIME_XML
	case "tar.gz":
		return mime.MIME_TGZ
	case "zip":
		return mime.MIME_GZIP
	}
	return mime.MIME_TGZ
}

/*
func IsMimeTypeSupported(mimeType string) bool {
	switch mimeType {
	case mime.MIME_JAR, mime.MIME_JSON, mime.MIME_XML, mime.MIME_TGZ, mime.MIME_GZIP:
		return true
	}
	return false
}
*/

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

func IsArtifact(fileName string) bool {
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
