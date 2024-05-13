package mvn

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/mimeutils"
)

// Coordinates holds the typical Maven coordinates groupId, artifactId, version. Optional also classifier and extension.
// https://maven.apache.org/ref/3.9.6/maven-core/artifact-handlers.html
type Coordinates struct {
	// GroupId of the Maven (mvn) artifact.
	GroupId string `json:"groupId"`
	// ArtifactId of the Maven (mvn) artifact.
	ArtifactId string `json:"artifactId"`
	// Version of the Maven (mvn) artifact.
	Version string `json:"version"`
	// Classifier of the Maven (mvn) artifact.
	Classifier string `json:"classifier"`
	// Extension of the Maven (mvn) artifact.
	Extension string `json:"extension"`
}

// GAV returns the GAV coordinates of the Maven Coordinates.
func (c *Coordinates) GAV() string {
	return c.GroupId + ":" + c.ArtifactId + ":" + c.Version
}

// String returns the Coordinates as a string (GroupId:ArtifactId:Version:Classifier:Extension).
func (c *Coordinates) String() string {
	return c.GroupId + ":" + c.ArtifactId + ":" + c.Version + ":" + c.Classifier + ":" + c.Extension
}

// GavPath returns the Maven repository path.
func (c *Coordinates) GavPath() string {
	return c.GroupPath() + "/" + c.ArtifactId + "/" + c.Version
}

// FilePath returns the Maven Coordinates's GAV-name with classifier and extension.
// Which is equal to the URL-path of the artifact in the repository.
// Default extension is jar.
func (c *Coordinates) FilePath() string {
	path := c.GavPath() + "/" + c.FileNamePrefix()
	if c.Classifier != "" {
		path += "-" + c.Classifier
	}
	if c.Extension != "" {
		path += "." + c.Extension
	} else {
		path += ".jar"
	}
	return path
}

func (c *Coordinates) Url(baseUrl string) string {
	return baseUrl + "/" + c.FilePath()
}

// GroupPath returns GroupId with `/` instead of `.`.
func (c *Coordinates) GroupPath() string {
	return strings.ReplaceAll(c.GroupId, ".", "/")
}

func (c *Coordinates) FileNamePrefix() string {
	return c.ArtifactId + "-" + c.Version
}

// Purl returns the Package URL of the Maven Coordinates.
func (c *Coordinates) Purl() string {
	return "pkg:maven/" + c.GroupId + "/" + c.ArtifactId + "@" + c.Version
}

// SetClassifierExtensionBy extracts the classifier and extension from the filename (without any path prefix).
func (c *Coordinates) SetClassifierExtensionBy(filename string) error {
	s := strings.TrimPrefix(path.Base(filename), c.FileNamePrefix())
	if strings.HasPrefix(s, "-") {
		s = strings.TrimPrefix(s, "-")
		i := strings.Index(s, ".")
		if i < 0 {
			return fmt.Errorf("no extension after classifier found in filename: %s", filename)
		}
		c.Classifier = s[:i]
		s = strings.TrimPrefix(s, c.Classifier)
	} else {
		c.Classifier = ""
	}
	c.Extension = strings.TrimPrefix(s, ".")
	return nil
}

// MimeType returns the MIME type of the Maven Coordinates based on the file extension.
// Default is application/x-tgz.
func (c *Coordinates) MimeType() string {
	m := mimeutils.TypeByExtension("." + c.Extension)
	if m != "" {
		return m
	}
	return mime.MIME_TGZ
}

// Copy creates a new Coordinates with the same values.
func (c *Coordinates) Copy() *Coordinates {
	return &Coordinates{
		GroupId:    c.GroupId,
		ArtifactId: c.ArtifactId,
		Version:    c.Version,
		Classifier: c.Classifier,
		Extension:  c.Extension,
	}
}

// Parse creates an Coordinates from it's serialized form (see Coordinates.String).
func Parse(serializedArtifact string) (*Coordinates, error) {
	parts := strings.Split(serializedArtifact, ":")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid artifact string: %s", serializedArtifact)
	}
	artifact := &Coordinates{
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
	return artifact, nil
}

// IsResource returns true if the filename is not a checksum or signature file.
func IsResource(fileName string) bool {
	switch filepath.Ext(fileName) {
	case ".asc", ".md5", ".sha1", ".sha256", ".sha512":
		return false
	default:
		return true
	}
}
