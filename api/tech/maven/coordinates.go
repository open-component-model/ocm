package maven

import (
	"crypto"
	"fmt"
	"mime"
	"path"
	"path/filepath"
	"strings"

	. "github.com/mandelsoft/goutils/regexutils"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	ocmmime "ocm.software/ocm/api/utils/mime"
)

type CoordinateOption = optionutils.Option[*Coordinates]

type WithClassifier string

func WithOptionalClassifier(c *string) CoordinateOption {
	if c != nil {
		return WithClassifier(*c)
	}
	return &optionutils.NoOption[*Coordinates]{}
}

func (o WithClassifier) ApplyTo(c *Coordinates) {
	c.Classifier = generics.PointerTo(string(o))
}

type WithExtension string

func WithOptionalExtension(e *string) CoordinateOption {
	if e != nil {
		return WithExtension(*e)
	}
	return &optionutils.NoOption[*Coordinates]{}
}

func (o WithExtension) ApplyTo(c *Coordinates) {
	c.Extension = generics.PointerTo(string(o))
}

type WithMediaType string

func WithOptionalMediaType(mt *string) CoordinateOption {
	if mt != nil {
		return WithMediaType(*mt)
	}
	return &optionutils.NoOption[*Coordinates]{}
}

func (o WithMediaType) ApplyTo(c *Coordinates) {
	c.MediaType = generics.PointerTo(string(o))
}

type FileCoordinates struct {
	// Classifier of the Maven artifact.
	Classifier *string `json:"classifier,omitempty"`
	// Extension of the Maven artifact.
	Extension *string `json:"extension,omitempty"`
	// MediaType of the files.
	MediaType *string `json:"mediaType,omitempty"`
}

// IsPackage returns true if the complete GAV content is addressed.
func (c *FileCoordinates) IsPackage() bool {
	return c.Classifier == nil && c.Extension == nil
}

// IsFile returns true if a dedicated single file is addressed.
func (c *FileCoordinates) IsFile() bool {
	return c.Classifier != nil && c.Extension != nil
}

// IsFileSet returns true if a file pattern is specified (and therefore, potentially multiple files are addressed).
func (c *FileCoordinates) IsFileSet() bool {
	return c.IsPackage() || !c.IsFile()
}

// MimeType returns the MIME type of the Maven Coordinates based on the file extension.
// Default is application/x-tgz.
func (c *FileCoordinates) MimeType() string {
	if c.Extension != nil && c.Classifier != nil {
		if c.MediaType != nil {
			return *c.MediaType
		}

		m := mime.TypeByExtension("." + optionutils.AsValue(c.Extension))
		if m != "" {
			return m
		}
		return ocmmime.MIME_OCTET
	}
	return ocmmime.MIME_TGZ
}

type PackageCoordinates struct {
	// GroupId of the Maven artifact.
	GroupId string `json:"groupId"`
	// ArtifactId of the Maven artifact.
	ArtifactId string `json:"artifactId"`
	// Version of the Maven artifact.
	Version string `json:"version"`
}

// GAV returns the GAV coordinates of the Maven Coordinates.
func (c *PackageCoordinates) GAV() string {
	return c.GroupId + ":" + c.ArtifactId + ":" + c.Version
}

func (c *PackageCoordinates) String() string {
	return c.GAV()
}

// GavPath returns the Maven repository path.
func (c *PackageCoordinates) GavPath() string {
	return c.GroupPath() + "/" + c.ArtifactId + "/" + c.Version
}

func (c *PackageCoordinates) GavLocation(repo *Repository) *Location {
	return repo.AddPath(c.GavPath())
}

// GroupPath returns GroupId with `/` instead of `.`.
func (c *PackageCoordinates) GroupPath() string {
	return strings.ReplaceAll(c.GroupId, ".", "/")
}

func (c *PackageCoordinates) FileNamePrefix() string {
	return c.ArtifactId + "-" + c.Version
}

// Purl returns the Package URL of the Maven Coordinates.
func (c *PackageCoordinates) Purl() string {
	return "pkg:maven/" + c.GroupId + "/" + c.ArtifactId + "@" + c.Version
}

// Coordinates holds the typical Maven coordinates groupId, artifactId, version. Optional also classifier and extension.
// https://maven.apache.org/ref/3.9.6/maven-core/artifact-handlers.html
type Coordinates struct {
	PackageCoordinates `json:",inline"`
	FileCoordinates    `json:",inline"`
}

func NewCoordinates(groupId, artifactId, version string, opts ...CoordinateOption) *Coordinates {
	c := &Coordinates{
		PackageCoordinates: PackageCoordinates{
			GroupId:    groupId,
			ArtifactId: artifactId,
			Version:    version,
		},
	}
	optionutils.ApplyOptions(c, opts...)
	return c
}

// String returns the Coordinates as a string (GroupId:ArtifactId:Version:WithClassifier:WithExtension).
func (c *Coordinates) String() string {
	return c.GroupId + ":" + c.ArtifactId + ":" + c.Version + ":" + optionutils.AsValue(c.Classifier) + ":" + optionutils.AsValue(c.Extension)
}

func (c *Coordinates) FileName() string {
	file := c.FileNamePrefix()
	if optionutils.AsValue(c.Classifier) != "" {
		file += "-" + *c.Classifier
	}
	if optionutils.AsValue(c.Extension) != "" {
		file += "." + *c.Extension
	} else {
		file += ".jar"
	}
	return file
}

// FilePath returns the Maven Coordinates's GAV-name with classifier and extension.
// Which is equal to the URL-path of the artifact in the repository.
// Default extension is jar.
func (c *Coordinates) FilePath() string {
	return c.GavPath() + "/" + c.FileName()
}

func (c *Coordinates) Location(repo *Repository) *Location {
	return repo.AddPath(c.FilePath())
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
		c.Classifier = generics.PointerTo(s[:i])
		s = strings.TrimPrefix(s, optionutils.AsValue(c.Classifier))
	} else {
		c.Classifier = generics.PointerTo("")
	}
	c.Extension = generics.PointerTo(strings.TrimPrefix(s, "."))
	return nil
}

// Copy creates a new Coordinates with the same values.
func (c *Coordinates) Copy() *Coordinates {
	return generics.PointerTo(*c)
}

func (c *Coordinates) FilterFileMap(fileMap map[string]crypto.Hash) map[string]crypto.Hash {
	if c.Classifier == nil && c.Extension == nil {
		return fileMap
	}
	exp := Literal(c.ArtifactId + "-" + c.Version)
	if optionutils.AsValue(c.Classifier) != "" {
		exp = Sequence(exp, Literal("-"+*c.Classifier))
	}
	if optionutils.AsValue(c.Extension) != "" {
		if c.Classifier == nil {
			exp = Sequence(exp, Optional(Literal("-"), Match(".+")))
		}
		exp = Sequence(exp, Literal("."+*c.Extension))
	} else {
		exp = Sequence(exp, Literal("."), Match(".*"))
	}
	exp = Anchored(exp)
	for file := range fileMap {
		if !exp.MatchString(file) {
			delete(fileMap, file)
		}
	}
	return fileMap
}

// Parse creates a Coordinates from its serialized form (see Coordinates.String).
func Parse(serializedArtifact string) (*Coordinates, error) {
	parts := strings.Split(serializedArtifact, ":")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid coordination string: %q", serializedArtifact)
	}
	coords := NewCoordinates(parts[0], parts[1], parts[2])
	if len(parts) >= 4 {
		coords.Classifier = generics.PointerTo(parts[3])
	}
	if len(parts) >= 5 {
		coords.Extension = generics.PointerTo(parts[4])
	}
	return coords, nil
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
